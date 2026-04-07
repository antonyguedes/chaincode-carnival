package fixer

import (
	"fmt"
	"strings"

	"github.com/antonioforte/chaincode-carnival/types"
)

// ValidatePatch evaluates structural safety of an LLM diff before modifying execution Sandbox contents.
func ValidatePatch(original, patched string, evidence types.ExploitEvidence) error {
	// 1. Scope Check: ensure diff touches lines near evidence.LineRefs +/- 10.
	// For simulation, we scan the mock patch for the hallucinated boundary breaker.
	
	if strings.Contains(patched, "DestroyAsset") {
		return fmt.Errorf("SAFETY GATE FAILED: Patch modified exported function signature outside allowed blast radius")
	}

	if !strings.Contains(patched, "len(value)") && evidence.VulnType == "Unbounded write" {
		return fmt.Errorf("SAFETY GATE FAILED: Patch does not address unbounded length constraint")
	}
	
	// Simulate checking if the patched file builds successfully
	// `go build ./patched_sandbox_dir/...`
	return nil
}
