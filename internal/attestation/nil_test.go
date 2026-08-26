package attestation

import (
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
)

func TestAttestationNilParamsNoPanic(t *testing.T) {
	a := Build(domain.Execution{ID: "ex-1", ActionKey: digest.OfString("k")},
		domain.BuildDefinition{ID: "def-1", ProjectID: "p-1", ToolchainID: "tc"},
		digest.OfString("in"), digest.OfString("out"), "")
	if a.Parameters == nil {
		t.Fatal("attestation parameters must not be nil even without executor version")
	}
}

func TestCompareEmptyAttestationSafe(t *testing.T) {
	b := domain.Attestation{ActionKey: digest.OfString("k"), InputRoot: digest.OfString("in"),
		OutputRoot: digest.OfString("out"), ExecutorVersion: "sim-executor/1"}
	diffs := Compare(domain.Attestation{}, b)
	if len(diffs) == 0 {
		t.Fatal("expected differences between zero and populated attestation")
	}
}
