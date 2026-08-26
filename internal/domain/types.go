package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"example.com/reproducible-build-farm/pkg/digest"
)

type ExecutionState string

const (
	StateQueued    ExecutionState = "queued"
	StateRunning   ExecutionState = "running"
	StateSucceeded ExecutionState = "succeeded"
	StateFailed    ExecutionState = "failed"
	StateCanceled  ExecutionState = "canceled"
)

type Project struct {
	ID, Name, Owner string
	CreatedAt       time.Time
	Version         int64
}
type Toolchain struct {
	ID, Name, Version, Platform string
	Digest                      digest.Digest
	Env                         map[string]string
}
type Input struct {
	Path   string
	Digest digest.Digest
	Size   int64
}
type Step struct {
	ID             string            `json:"id"`
	Inputs         []Input           `json:"inputs"`
	Dependencies   []string          `json:"dependencies"`
	Args           []string          `json:"args"`
	Env            map[string]string `json:"env"`
	Platform       string            `json:"platform"`
	Outputs        []string          `json:"outputs"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	Network        bool              `json:"network"`
}
type BuildDefinition struct {
	ID, ProjectID, Name string
	ToolchainID         string
	Steps               []Step
	Resource            ResourceBudget
	CreatedAt           time.Time
}
type ResourceBudget struct {
	CPU            int
	MemoryMB       int
	TimeoutSeconds int
	RetryLimit     int
}
type Execution struct {
	ID, DefinitionID, IdempotencyKey string
	State                            ExecutionState
	ActionKey                        digest.Digest
	ResultDigest                     digest.Digest
	AttestationID                    string
	Error                            string
	CreatedAt, StartedAt, FinishedAt time.Time
	Attempt                          int
}
type Worker struct {
	ID, Platform, Version string
	Capacity              ResourceBudget
	Busy                  bool
	LastHeartbeat         time.Time
	LeaseID               string
}
type CacheEntry struct {
	ActionKey            digest.Digest
	ResultDigest         digest.Digest
	OutputManifest       map[string]digest.Digest
	LogsDigest           digest.Digest
	CreatedAt, ExpiresAt time.Time
	Negative             bool
}
type Attestation struct {
	ID, ExecutionID                  string
	ActionKey, InputRoot, OutputRoot digest.Digest
	ToolchainDigest                  digest.Digest
	Parameters                       map[string]string
	ExecutorVersion                  string
	StartedAt, FinishedAt            time.Time
	Signature                        digest.Digest
}
type Lease struct {
	ID, ExecutionID, WorkerID string
	Version                   int64
	ExpiresAt                 time.Time
}

func (s Step) Canonical() string {
	deps := append([]string(nil), s.Dependencies...)
	sort.Strings(deps)
	args := append([]string(nil), s.Args...)
	sort.Strings(args)
	env := make([]string, 0, len(s.Env))
	for k, v := range s.Env {
		env = append(env, k+"="+v)
	}
	sort.Strings(env)
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d|%t", s.ID, strings.Join(deps, ","), strings.Join(args, ","), strings.Join(env, ";"), s.Platform, s.TimeoutSeconds, s.Network)
}
func (e Execution) Terminal() bool {
	return e.State == StateSucceeded || e.State == StateFailed || e.State == StateCanceled
}
