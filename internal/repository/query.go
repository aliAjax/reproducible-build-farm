package repository

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"sort"
	"strings"
)

type ExecutionFilter struct {
	State     domain.ExecutionState
	ProjectID string
	Limit     int
}

func (m *Memory) QueryExecutions(ctx context.Context, f ExecutionFilter) []domain.Execution {
	all := m.ListExecutions(ctx)
	out := []domain.Execution{}
	for _, e := range all {
		if f.State != "" && e.State != f.State {
			continue
		}
		if f.ProjectID != "" {
			d, err := m.GetDefinition(ctx, e.DefinitionID)
			if err != nil || d.ProjectID != f.ProjectID {
				continue
			}
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if f.Limit > 0 && len(out) > f.Limit {
		out = out[:f.Limit]
	}
	return out
}
func (m *Memory) DeleteExpiredCache(nowUnix int64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for k, v := range m.cache {
		if !v.ExpiresAt.IsZero() && v.ExpiresAt.Unix() <= nowUnix {
			delete(m.cache, k)
			n++
		}
	}
	return n
}
func NormalizeIdempotency(value string) string { return strings.TrimSpace(value) }
