package dsl

import (
	"encoding/json"
	"example.com/reproducible-build-farm/internal/domain"
	"fmt"
	"sort"
	"strings"
)

func Canonical(doc Document) ([]byte, error) {
	steps := append([]domain.Step(nil), doc.Steps...)
	sort.Slice(steps, func(i, j int) bool { return steps[i].ID < steps[j].ID })
	for i := range steps {
		sort.Strings(steps[i].Dependencies)
		sort.Strings(steps[i].Args)
		env := map[string]string{}
		keys := make([]string, 0, len(steps[i].Env))
		for k := range steps[i].Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			env[k] = strings.TrimSpace(steps[i].Env[k])
		}
		steps[i].Env = env
	}
	doc.Steps = steps
	return json.Marshal(doc)
}
func ValidateEnv(env map[string]string) error {
	for k, v := range env {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("empty environment key")
		}
		if len(k) > 128 || len(v) > 4096 {
			return fmt.Errorf("environment entry too large")
		}
		if strings.ContainsAny(k, "= \t\n") {
			return fmt.Errorf("invalid environment key %q", k)
		}
	}
	return nil
}
func AllowedArgument(arg string) bool {
	if strings.ContainsAny(arg, ";&|<>`$(){}\n\r") {
		return false
	}
	return len(arg) <= 1024
}
func ValidateArgs(args []string) error {
	for _, a := range args {
		if !AllowedArgument(a) {
			return fmt.Errorf("unsafe argument")
		}
	}
	return nil
}
func NormalizePath(p string) string {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	p = strings.TrimPrefix(p, "./")
	return p
}
func CheckOutputs(outputs []string) error {
	seen := map[string]bool{}
	for _, o := range outputs {
		n := NormalizePath(o)
		if n == "" || strings.HasPrefix(n, "/") || strings.HasPrefix(n, "../") || seen[n] {
			return fmt.Errorf("invalid output %q", o)
		}
		seen[n] = true
	}
	return nil
}
