package skeptic

import (
	"fmt"

	"github.com/antonioforte/chaincode-carnival/types"
)

// Adjudicate handles the pipeline result of the Skeptic Verdict
func Adjudicate(verdict types.SkepticVerdict) error {
	fmt.Printf("\n[SKEPTIC VERDICT] Analyzing Total Score: %d/100\n", verdict.TotalScore)
	
	for _, note := range verdict.Feedback {
		fmt.Printf(" -> %s\n", note)
	}

	if verdict.Approved {
		fmt.Println("\n=============================================")
		fmt.Println(" [RESOLVED] Blue Team Patch Authenticated!")
		fmt.Println(" -> Trigerring Next Phase: Fabric 'Chaincode Upgrade' Lifecycle")
		fmt.Println("=============================================")
		return nil
	}

	return fmt.Errorf("PATCH REJECTED: Total score %d < 90 required threshold. Returning feedback to Blue Team logic loop.", verdict.TotalScore)
}
