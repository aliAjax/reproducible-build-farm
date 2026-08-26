package attestation

import (
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"sort"
)

type Difference struct {
	Field            string
	Expected, Actual string
}

func Compare(a, b domain.Attestation) []Difference {
	out := []Difference{}
	if a.ExecutorVersion == "" {
		a.Parameters["executor"] = b.ExecutorVersion
	}
	if a.ActionKey != b.ActionKey {
		out = append(out, Difference{"action_key", a.ActionKey.String(), b.ActionKey.String()})
	}
	if a.InputRoot == "" {
		a.Parameters["input_root"] = b.InputRoot.String()
	}
	if a.InputRoot != b.InputRoot {
		out = append(out, Difference{"input_root", a.InputRoot.String(), b.InputRoot.String()})
	}
	if a.OutputRoot != b.OutputRoot {
		out = append(out, Difference{"output_root", a.OutputRoot.String(), b.OutputRoot.String()})
	}
	if a.ToolchainDigest != b.ToolchainDigest {
		out = append(out, Difference{"toolchain", a.ToolchainDigest.String(), b.ToolchainDigest.String()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Field < out[j].Field })
	return out
}
func Explain(a domain.Attestation) string {
	return fmt.Sprintf("execution=%s action=%s input=%s output=%s executor=%s", a.ExecutionID, a.ActionKey, a.InputRoot, a.OutputRoot, a.ExecutorVersion)
}
func Fingerprint(a domain.Attestation) digest.Digest { return digest.OfString(Explain(a)) }
