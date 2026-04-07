package fixer

import (
	"fmt"
	"time"

	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/types"
)

const maxRounds = 5

type PatchResult struct {
	Patch  string
	Rounds int
}

type ExploitRunner func(patchedCode string, evidence types.ExploitEvidence) types.RetestResult
type AuditLogger func(event types.AuditEvent)

func FixerLoop(evidence types.ExploitEvidence, src string, exploitRunner ExploitRunner, auditLog AuditLogger, provider providers.LLMProvider) (PatchResult, error) {
	fmt.Printf("\n[BLUE TEAM] Commencing Adversarial Loop for Session: %s\n", evidence.SessionID)
	
	for round := 0; round < maxRounds; round++ {
		fmt.Printf("\n=> ROUND %d\n", round)
		
		patch, err := GeneratePatch(evidence, src, "", round, provider)
		if err != nil {
			fmt.Printf(" [!] Patch generation collapsed: %v\n", err)
			time.Sleep(4 * time.Second) // API Rate limit buffer before next round
			continue
		}
		
		fmt.Printf(" [~] ValidatePatch Executing...\n")
		if err := ValidatePatch(src, patch, evidence); err != nil {
			fmt.Printf(" [!] Validation Rejected: %v.\n     -> Sending specific trace failure map back to LLM Constraints.\n", err)
			// inner loop would re-prompt LLM here up to 3 times before moving up. Proceeding round counter forward for simulation.
			continue 
		}

		fmt.Printf(" [+] Validation Passed! Pinging Red Team for Retest in Sandbox...\n")
		retestResult := exploitRunner(patch, evidence)

		auditLog(types.AuditEvent{
			EventType: "PATCH_ROUND",
			SessionID: evidence.SessionID,
			Payload:   fmt.Sprintf(`{"round": %d, "patch": "%s", "closed": %t}`, round, patch, retestResult.ExploitClosed),
		})

		if retestResult.ExploitClosed && len(retestResult.NewVulns) == 0 {
			fmt.Printf(" [SUCCESS] Blue Team Exploit defensively closed after %d patch rounds!\n", round+1)
			return PatchResult{Patch: patch, Rounds: round + 1}, nil
		}

		if len(retestResult.NewVulns) > 0 {
			fmt.Printf(" [!] Patch caused architectural regression vulnerabilities.\n")
		} else {
		    fmt.Printf(" [!] Exploit remains active with current patch. Escalating payload constraint structure variables...\n")
		}
	}
	
	// Write escalation block safely to audit trail
	auditLog(types.AuditEvent{
		EventType: "HUMAN_REVIEW_REQUIRED",
		SessionID: evidence.SessionID,
		Payload:   "Max rounds reached without vulnerability resolution",
	})
	
	return PatchResult{}, fmt.Errorf("escalated: max rounds reached")
}
