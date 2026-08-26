package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var identifier = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)

func ValidateIdentifier(value string) error {
	if !identifier.MatchString(value) {
		return fmt.Errorf("identifier must match %s", identifier.String())
	}
	return nil
}
func (p Project) Validate() error {
	if err := ValidateIdentifier(p.ID); err != nil {
		return err
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("project name is required")
	}
	if len(p.Name) > 200 {
		return fmt.Errorf("project name too long")
	}
	return nil
}
func (t Toolchain) Validate() error {
	if err := ValidateIdentifier(t.ID); err != nil {
		return err
	}
	if t.Digest.Valid() == false {
		return fmt.Errorf("toolchain digest invalid")
	}
	if t.Platform == "" {
		return fmt.Errorf("platform required")
	}
	return nil
}
func (r ResourceBudget) Validate() error {
	if r.CPU < 0 || r.MemoryMB < 0 || r.TimeoutSeconds < 0 || r.RetryLimit < 0 {
		return fmt.Errorf("resource values cannot be negative")
	}
	if r.CPU > 1024 || r.MemoryMB > 1048576 {
		return fmt.Errorf("resource values exceed limits")
	}
	return nil
}
func (s Step) Validate() error {
	if err := ValidateIdentifier(s.ID); err != nil {
		return err
	}
	if len(s.Args) > 128 {
		return fmt.Errorf("too many arguments")
	}
	if len(s.Outputs) > 256 {
		return fmt.Errorf("too many outputs")
	}
	for _, o := range s.Outputs {
		if strings.HasPrefix(o, "/") || strings.Contains(o, "..") {
			return fmt.Errorf("unsafe output path")
		}
	}
	return nil
}
func (d BuildDefinition) Validate() error {
	if err := ValidateIdentifier(d.ID); err != nil {
		return err
	}
	if err := ValidateIdentifier(d.ProjectID); err != nil {
		return err
	}
	if len(d.Steps) == 0 {
		return fmt.Errorf("at least one step required")
	}
	if err := d.Resource.Validate(); err != nil {
		return err
	}
	for _, s := range d.Steps {
		if err := s.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func (e Execution) Validate() error {
	if err := ValidateIdentifier(e.ID); err != nil {
		return err
	}
	if e.DefinitionID == "" {
		return fmt.Errorf("definition id required")
	}
	switch e.State {
	case StateQueued, StateRunning, StateSucceeded, StateFailed, StateCanceled:
	default:
		return fmt.Errorf("unknown execution state")
	}
	return nil
}
func CanTransition(from, to ExecutionState) bool {
	switch from {
	case StateQueued:
		return to == StateRunning || to == StateCanceled
	case StateRunning:
		return to == StateSucceeded || to == StateFailed || to == StateCanceled
	default:
		return false
	}
}
