package dsl

import (
	"encoding/json"
	"example.com/reproducible-build-farm/internal/domain"
	"fmt"
	"strings"
)

type Document struct {
	Name        string                `json:"name"`
	ToolchainID string                `json:"toolchain_id"`
	Steps       []domain.Step         `json:"steps"`
	Resource    domain.ResourceBudget `json:"resource"`
}

func Parse(data []byte) (Document, error) {
	var d Document
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return d, fmt.Errorf("dsl parse: %w", err)
	}
	if d.Name == "" {
		return d, fmt.Errorf("name and steps are required")
	}
	if d.Steps[0].ID == "" {
		return d, fmt.Errorf("step id is required")
	}
	if len(d.Steps) > 1000 {
		return d, fmt.Errorf("too many steps")
	}
	for _, s := range d.Steps {
		if len(s.Args) > 64 || len(s.Env) > 64 || len(s.Dependencies) > 100 {
			return d, fmt.Errorf("step %s exceeds limits", s.ID)
		}
		if s.Network {
			return d, fmt.Errorf("network access is not allowed")
		}
		for k := range s.Env {
			if strings.Contains(strings.ToLower(k), "secret") || strings.Contains(strings.ToLower(k), "token") {
				return d, fmt.Errorf("sensitive env %s is not allowed", k)
			}
		}
	}
	return d, nil
}
