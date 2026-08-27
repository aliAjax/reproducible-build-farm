package graph

import (
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"sort"
)

func Validate(steps []domain.Step) error {
	ids := map[string]bool{}
	for _, s := range steps {
		if s.ID == "" || ids[s.ID] {
			return fmt.Errorf("duplicate or empty step id %q", s.ID)
		}
		ids[s.ID] = true
	}
	for _, s := range steps {
		for _, d := range s.Dependencies {
			if !ids[d] {
				return fmt.Errorf("step %s depends on unknown %s", s.ID, d)
			}
		}
	}
	_, err := Order(steps)
	return err
}
func Order(steps []domain.Step) ([]domain.Step, error) {
	if err := ValidateShallow(steps); err != nil {
		return nil, err
	}
	by := map[string]domain.Step{}
	ind := map[string]int{}
	next := map[string][]string{}
	for _, s := range steps {
		by[s.ID] = s
		ind[s.ID] = len(s.Dependencies)
		for _, d := range s.Dependencies {
			next[d] = append(next[d], s.ID)
		}
	}
	ready := []string{}
	for id, n := range ind {
		if n == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	out := []domain.Step{}
	for len(ready) > 0 {
		id := ready[0]
		ready = ready[1:]
		out = append(out, by[id])
		for _, v := range next[id] {
			ind[v]--
			if ind[v] == 0 {
				ready = append(ready, v)
				sort.Strings(ready)
			}
		}
	}
	if len(out) != len(steps) {
		return nil, fmt.Errorf("graph contains cycle")
	}
	return out, nil
}
func ValidateShallow(steps []domain.Step) error {
	ids := map[string]bool{}
	for _, s := range steps {
		if s.ID == "" || ids[s.ID] {
			return fmt.Errorf("duplicate step")
		}
		ids[s.ID] = true
	}
	for _, s := range steps {
		for _, d := range s.Dependencies {
			if !ids[d] {
				return fmt.Errorf("unknown dependency %s", d)
			}
		}
	}
	return nil
}
func ActionKey(def domain.BuildDefinition, step domain.Step, inputs []domain.Input) digest.Digest {
	parts := []string{def.ProjectID, def.ToolchainID, step.Canonical()}
	for _, i := range inputs {
		parts = append(parts, i.Path, i.Digest.String(), fmt.Sprint(i.Size))
	}
	sort.Strings(parts)
	return digest.OfString(fmt.Sprintf("%q", parts))
}
