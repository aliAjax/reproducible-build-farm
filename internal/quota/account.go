package quota

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"fmt"
	"sync"
)

type Account struct {
	mu       sync.Mutex
	name     string
	limit    domain.ResourceBudget
	reserved map[string]domain.ResourceBudget
}

func NewAccount(name string, limit domain.ResourceBudget) *Account {
	return &Account{name: name, limit: limit, reserved: map[string]domain.ResourceBudget{}}
}
func (a *Account) Reserve(_ context.Context, id string, r domain.ResourceBudget) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.reserved[id]; ok {
		return fmt.Errorf("reservation exists")
	}
	var cpu, mem int
	for _, v := range a.reserved {
		cpu += v.CPU
		mem += v.MemoryMB
	}
	if a.limit.CPU > 0 && cpu+r.CPU > a.limit.CPU {
		return fmt.Errorf("cpu quota exceeded")
	}
	if a.limit.MemoryMB > 0 && mem+r.MemoryMB > a.limit.MemoryMB {
		return fmt.Errorf("memory quota exceeded")
	}
	a.reserved[id] = r
	return nil
}
func (a *Account) Release(id string) { a.mu.Lock(); defer a.mu.Unlock(); delete(a.reserved, id) }
func (a *Account) Usage() domain.ResourceBudget {
	a.mu.Lock()
	defer a.mu.Unlock()
	var out domain.ResourceBudget
	for _, v := range a.reserved {
		out.CPU += v.CPU
		out.MemoryMB += v.MemoryMB
	}
	return out
}
