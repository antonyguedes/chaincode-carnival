package fixer

import (
	"fmt"
	"strings"

	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/types"
)

// GeneratePatch dynamically asks the LLM to generate a unified diff for the broken chaincode.
func GeneratePatch(evidence types.ExploitEvidence, src string, additionalConstraints string, round int, provider providers.LLMProvider) (string, error) {
	prompt := fmt.Sprintf(`You are an expert Blue Team Golang Engineer. The Red Team has successfully breached the Sandbox using the following vulnerability:
Vulnerability: %s
Offending Code Lines: %v
Chaincode Source:
---
%s
---

Additional Round Constraints: %s

You MUST return a strict Unified Diff format string representing the minimal patch needed to secure the software. 
DO NOT modify any exported function signatures or write changes far outside the indicated lines. 
Output ONLY the @@ diff block and nothing else.
Diff:`, evidence.VulnType, evidence.LineRefs, src, additionalConstraints)

	respRaw, err := provider.Query(prompt)
	if err != nil {
		return "", fmt.Errorf("[%s] Fixer patch generation failed: %w", provider.Name(), err)
	}

	// Try extracting standard diff block
	start := strings.Index(respRaw, "@@")
	if start == -1 {
		return "", fmt.Errorf("[%s] Failed to extract Unified Diff block. Raw: %s", provider.Name(), respRaw)
	}
	match := respRaw[start:]
	match = strings.ReplaceAll(match, "```diff", "")
	match = strings.ReplaceAll(match, "```", "")
	match = strings.TrimSpace(match)

	return match, nil
}
