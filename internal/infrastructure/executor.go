package infrastructure

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Result struct {
	OutputManifest    map[string]digest.Digest
	Logs              string
	ResultDigest      digest.Digest
	Started, Finished time.Time
}
type Executor interface {
	Execute(context.Context, domain.Step, []domain.Input) (Result, error)
	Version() string
}
type SimulatedExecutor struct{ VersionString string }

func NewSimulatedExecutor() *SimulatedExecutor {
	return &SimulatedExecutor{VersionString: "sim-executor/1"}
}
func (e *SimulatedExecutor) Version() string { return e.VersionString }
func (e *SimulatedExecutor) Execute(ctx context.Context, s domain.Step, inputs []domain.Input) (Result, error) {
	start := time.Now().UTC()
	names := []string{}
	for _, i := range inputs {
		names = append(names, i.Path+":"+i.Digest.String())
	}
	sort.Strings(names)
	payload := strings.Join([]string{s.Canonical(), strings.Join(names, ";")}, "|")
	rd := digest.OfString(payload)
	out := map[string]digest.Digest{}
	for _, p := range s.Outputs {
		out[p] = digest.OfString(payload + "|" + p)
	}
	return Result{OutputManifest: out, Logs: fmt.Sprintf("simulated step %s outputs=%d", s.ID, len(out)), ResultDigest: rd, Started: start, Finished: time.Now().UTC()}, nil
}
