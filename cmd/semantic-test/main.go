package main

import (
	"encoding/json"
	"fmt"

	"github.com/antonioforte/chaincode-carnival/agents/analyzer"
	"github.com/antonioforte/chaincode-carnival/agents/analyzer/llm"
	"github.com/antonioforte/chaincode-carnival/types"
)

func main() {
	fmt.Println("=== Analyzer: LLM Semantic & Merging Validation ===")

	// 1. Emulate an AST Extraction providing valid function names
	astFunctions := []string{"GetAssetHistory", "CreateAsset", "TransferAsset"}

	// 2. Emulate an incoming LLM response processing a codebase
	rawLLMFindings := llm.AnalyzeSemantics("... mocked source code ...")

	fmt.Println("\n[1] Checking Raw LLM Output...")
	for _, f := range rawLLMFindings {
		fmt.Printf(" - Found %s targeting %s (Confidence: %s)\n", f.RuleID, f.FuncName, f.Confidence)
	}

	// 3. Apply the Hallucination Validator filter
	fmt.Println("\n[2] Applying Hallucination Validator against valid AST blocks...")
	validatedLLMFindings := llm.ValidateFindings(rawLLMFindings, astFunctions)
	for _, f := range validatedLLMFindings {
		fmt.Printf(" - Validated %s targeting %s (Confidence: %s)\n", f.RuleID, f.FuncName, f.Confidence)
	}

	// 4. Mock an AST rule return finding for context
	astFindings := []types.Finding{
		{
			RuleID:     "ACL-001",
			VulnType:   "Missing identity check",
			Severity:   "critical",
			FuncName:   "TransferAsset",
			Confidence: "certain",
			Source:     "ast",
		},
	}

	// 5. Build Final Report
	fmt.Println("\n[3] Generating merged AnalyzerReport priority queue...")
	finalReport := analyzer.MergeReports("test-chaincode", astFindings, validatedLLMFindings)
	
	out, _ := json.MarshalIndent(finalReport, "", "  ")
	fmt.Println(string(out))
}
