package attestation

import (
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"sort"
)

func Build(e domain.Execution, def domain.BuildDefinition, inputRoot, outputRoot digest.Digest, executor string) domain.Attestation {
	params := map[string]string{"definition": def.ID, "project": def.ProjectID}
	if executor != "" {
		params["executor"] = executor
	}
	keys := []string{}
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sig := digest.OfString(fmt.Sprintf("%s|%s|%s|%s|%s", e.ID, e.ActionKey, inputRoot, outputRoot, executor))
	return domain.Attestation{ID: "att-" + e.ID, ExecutionID: e.ID, ActionKey: e.ActionKey, InputRoot: inputRoot, OutputRoot: outputRoot, ToolchainDigest: digest.OfString(def.ToolchainID), Parameters: params, ExecutorVersion: executor, StartedAt: e.StartedAt, FinishedAt: e.FinishedAt, Signature: sig}
}
func Verify(a domain.Attestation) bool { return a.Signature.Valid() }
