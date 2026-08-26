package graph

import (
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"sort"
)

type Plan struct {
	Steps  []domain.Step
	Levels [][]string
	Root   digest.Digest
}

func BuildPlan(steps []domain.Step) (Plan, error) {
	sorted := make([]domain.Step, len(steps))
	copy(sorted, steps)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	ordered, err := Order(sorted)
	if err != nil {
		return Plan{}, err
	}
	levels := [][]string{}
	depth := map[string]int{}
	for _, s := range ordered {
		d := 0
		for _, dep := range s.Dependencies {
			if depth[dep]+1 > d {
				d = depth[dep] + 1
			}
		}
		depth[s.ID] = d
		for len(levels) <= d {
			levels = append(levels, []string{})
		}
		levels[d] = append(levels[d], s.ID)
	}
	for i := range levels {
		sort.Strings(levels[i])
	}
	parts := []string{}
	for _, s := range ordered {
		parts = append(parts, s.ID, s.Canonical())
	}
	return Plan{Steps: ordered, Levels: levels, Root: digest.OfString(fmt.Sprintf("%q", parts))}, nil
}
func Ready(steps []domain.Step, completed map[string]bool) []domain.Step {
	out := []domain.Step{}
	for _, s := range steps {
		if completed[s.ID] {
			continue
		}
		ok := true
		for _, d := range s.Dependencies {
			if !completed[d] {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func Dependents(steps []domain.Step, id string) []string {
	out := []string{}
	for _, s := range steps {
		for _, d := range s.Dependencies {
			if d == id {
				out = append(out, s.ID)
			}
		}
	}
	sort.Strings(out)
	return out
}
func CriticalPath(steps []domain.Step) int {
	depth := map[string]int{}
	ordered, _ := Order(steps)
	for _, s := range ordered {
		for _, d := range s.Dependencies {
			if depth[d]+1 > depth[s.ID] {
				depth[s.ID] = depth[d] + 1
			}
		}
	}
	max := 0
	for _, v := range depth {
		if v > max {
			max = v
		}
	}
	return max + 1
}
