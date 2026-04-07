package main

import (
	"fmt"

	"github.com/antonioforte/chaincode-carnival/agents/skeptic"
	"github.com/antonioforte/chaincode-carnival/types"
)

func main() {
	fmt.Println("=== Skeptic Agent Testing CLI ===")

	mockLogs := []types.AuditEvent{
		{EventType: "PATCH_ROUND", Payload: "mock payload round 1"},
	}

	fmt.Println("\n[Test 1] Passing an unidiomatic, bad patch...")
	badPatch := `@@ -10,6 +10,8 @@
 func TransferAsset(...) {
+ // TODO: check if this is right
+ var temp_hack bool = true
  return
 }`

	verdict1 := skeptic.EvaluatePatch(mockLogs, badPatch)
	err := skeptic.Adjudicate(verdict1)
	if err != nil {
		fmt.Printf("Outcome: %v\n", err)
	}

	fmt.Println("\n[Test 2] Passing a robust, idiomatic patch...")
	goodPatch := `@@ -66,6 +66,8 @@
 func CreateAsset(...) {
+ if len(value) > 1024 {
+     return fmt.Errorf("payload exceeds permitted max size")
+ }
  return
 }`

	verdict2 := skeptic.EvaluatePatch(mockLogs, goodPatch)
	err = skeptic.Adjudicate(verdict2)
	if err != nil {
		fmt.Printf("Outcome: %v\n", err)
	}
}
