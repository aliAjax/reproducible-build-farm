package infrastructure

import (
	"example.com/reproducible-build-farm/internal/domain"
	"fmt"
	"strings"
)

type ConstraintSet struct {
	Platform        string
	AllowedEnv      map[string]bool
	MaxArgs         int
	RequireApproval bool
}

func (c ConstraintSet) Check(s domain.Step) error {
	if c.Platform != "" && s.Platform != "" && c.Platform != s.Platform {
		return fmt.Errorf("platform mismatch")
	}
	if c.MaxArgs > 0 && len(s.Args) > c.MaxArgs {
		return fmt.Errorf("argument count exceeds limit")
	}
	for k := range s.Env {
		if !c.AllowedEnv[k] {
			return fmt.Errorf("environment %s is not allowlisted", k)
		}
	}
	for _, a := range s.Args {
		if strings.ContainsAny(a, ";&|<>$`\n") {
			return fmt.Errorf("unsafe argument")
		}
	}
	return nil
}
func CheckBudget(b domain.ResourceBudget) error {
	if b.CPU < 1 || b.MemoryMB < 64 {
		return fmt.Errorf("minimum budget is 1 cpu and 64MB")
	}
	if b.TimeoutSeconds > 86400 {
		return fmt.Errorf("timeout exceeds 24h")
	}
	return nil
}
func MergeEnv(base, overlay map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overlay {
		out[k] = v
	}
	return out
}
