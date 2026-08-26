package application

import (
	"context"
	"example.com/reproducible-build-farm/internal/attestation"
	"example.com/reproducible-build-farm/internal/cache"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/dsl"
	"example.com/reproducible-build-farm/internal/graph"
	"example.com/reproducible-build-farm/internal/infrastructure"
	"example.com/reproducible-build-farm/internal/quota"
	"example.com/reproducible-build-farm/internal/repository"
	"example.com/reproducible-build-farm/internal/worker"
	"example.com/reproducible-build-farm/pkg/clock"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"sync"
	"time"
)

type Service struct {
	Store    repository.Store
	Cache    cache.Remote
	Executor infrastructure.Executor
	Leases   *worker.Manager
	Quota    *quota.Manager
	Clock    clock.Clock
	mu       sync.Mutex
}

func New(s repository.Store, c cache.Remote, e infrastructure.Executor) *Service {
	return &Service{Store: s, Cache: c, Executor: e, Leases: worker.NewManager(s, 30*time.Second), Quota: quota.New(domain.ResourceBudget{CPU: 64, MemoryMB: 65536}), Clock: clock.Real{}}
}
func (s *Service) CreateDefinition(ctx context.Context, projectID, id string, data []byte) (domain.BuildDefinition, error) {
	doc, err := dsl.Parse(data)
	if err != nil {
		return domain.BuildDefinition{}, err
	}
	if err = graph.Validate(doc.Steps); err != nil {
		return domain.BuildDefinition{}, err
	}
	d := domain.BuildDefinition{ID: id, ProjectID: projectID, Name: doc.Name, ToolchainID: doc.ToolchainID, Steps: doc.Steps, Resource: doc.Resource, CreatedAt: s.Clock.Now()}
	return d, s.Store.SaveDefinition(ctx, d)
}
func (s *Service) Submit(ctx context.Context, id, defID, idem string) (domain.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idem != "" {
		if old, err := s.Store.FindExecutionByIdempotency(ctx, idem); err == nil {
			return old, nil
		}
	}
	d, err := s.Store.GetDefinition(ctx, defID)
	if err != nil {
		return domain.Execution{}, err
	}
	key := digest.OfString(fmt.Sprintf("%s|%s", d.ID, d.ToolchainID))
	e := domain.Execution{ID: id, DefinitionID: defID, IdempotencyKey: idem, State: domain.StateQueued, ActionKey: key, CreatedAt: s.Clock.Now()}
	if err = s.Store.SaveExecution(ctx, e); err != nil {
		return e, err
	}
	go s.run(context.Background(), e, d)
	return e, nil
}
func (s *Service) run(ctx context.Context, e domain.Execution, d domain.BuildDefinition) {
	_ = s.Quota.Reserve(ctx, d.Resource)
	defer s.Quota.Release(ctx, d.Resource)
	e.State = domain.StateRunning
	e.StartedAt = s.Clock.Now()
	_ = s.Store.SaveExecution(ctx, e)
	ordered, err := graph.Order(d.Steps)
	if err != nil {
		s.fail(ctx, e, err)
		return
	}
	inputs := []domain.Input{}
	outRoot := digest.OfString("")
	for _, step := range ordered {
		key := graph.ActionKey(d, step, inputs)
		if c, ok := s.Cache.Get(ctx, key); ok {
			e.ResultDigest = c.ResultDigest
			outRoot = c.ResultDigest
			continue
		}
		lease, err := s.Leases.Acquire(ctx, e.ID, d.Resource)
		if err != nil {
			s.fail(ctx, e, err)
			return
		}
		r, err := s.Executor.Execute(ctx, step, inputs)
		_ = s.Leases.Release(ctx, lease)
		if err != nil {
			s.fail(ctx, e, err)
			return
		}
		s.Cache.Put(ctx, domain.CacheEntry{ActionKey: key, ResultDigest: r.ResultDigest, OutputManifest: r.OutputManifest, LogsDigest: digest.OfString(r.Logs), CreatedAt: s.Clock.Now(), ExpiresAt: s.Clock.Now().Add(24 * time.Hour)})
		e.ResultDigest = r.ResultDigest
		outRoot = r.ResultDigest
	}
	e.State = domain.StateSucceeded
	e.FinishedAt = s.Clock.Now()
	a := attestation.Build(e, d, digest.OfString(d.ProjectID), outRoot, s.Executor.Version())
	e.AttestationID = a.ID
	_ = s.Store.SaveAttestation(ctx, a)
	_ = s.Store.SaveExecution(ctx, e)
}
func (s *Service) fail(ctx context.Context, e domain.Execution, err error) {
	e.State = domain.StateFailed
	e.Error = err.Error()
	e.FinishedAt = s.Clock.Now()
	_ = s.Store.SaveExecution(ctx, e)
}
func (s *Service) GetExecution(ctx context.Context, id string) (domain.Execution, error) {
	return s.Store.GetExecution(ctx, id)
}
func (s *Service) GetAttestation(ctx context.Context, id string) (domain.Attestation, error) {
	return s.Store.GetAttestation(ctx, id)
}
