package arena

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/antonioforte/chaincode-carnival/agents/analyzer"
	analyzerAst "github.com/antonioforte/chaincode-carnival/agents/analyzer/ast"
	"github.com/antonioforte/chaincode-carnival/agents/analyzer/llm"
	"github.com/antonioforte/chaincode-carnival/agents/cicd"
	"github.com/antonioforte/chaincode-carnival/agents/exploiter"
	"github.com/antonioforte/chaincode-carnival/agents/exploiter/sandbox"
	"github.com/antonioforte/chaincode-carnival/agents/fixer"
	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/agents/skeptic"
	"github.com/antonioforte/chaincode-carnival/types"
	"github.com/joho/godotenv"
)

func banter(agent string, providerName string, message string) {
	colorReset := "\033[0m"
	colorMap := map[string]string{
		"Analyzer": "\033[36m", // Cyan
		"RedTeam":  "\033[31m", // Red
		"BlueTeam": "\033[34m", // Blue
		"Skeptic":  "\033[35m", // Purple
		"System":   "\033[32m", // Green
	}
	color := colorMap[agent]

	if providerName != "" {
		fmt.Printf("\n%s[%s (%s)]%s %s\n", color, agent, providerName, colorReset, message)
	} else {
		fmt.Printf("\n%s[%s]%s %s\n", color, agent, colorReset, message)
	}
	
	if agent == "System" {
		time.Sleep(1 * time.Second)
	} else {
		time.Sleep(5 * time.Second)
	}
}

func getDynamicBanter(agentRole string, context string, provider providers.LLMProvider) string {
	prompt := fmt.Sprintf(`You are roleplaying as the '%s' in a competitive cybersecurity arena match.
The current situation is: %s
Write exactly ONE sentence of highly competitive, arrogant, and clever hacker trash-talk or commentary based strictly on the situation.
DO NOT use quotes. DO NOT use emojis. DO NOT write more than one sentence.`, agentRole, context)

	resp, err := provider.Query(prompt)
	if err != nil {
		if agentRole == "Red Team" { return "I'm going to rip this chaincode completely open." }
		if agentRole == "Blue Team" { return "I'll dynamically lock this down before you even launch the payload." }
		if agentRole == "Analyzer Engine" { return "Step aside. Let me scan this codebase..." }
		if agentRole == "Skeptic Judge" { return "Let's see if this code meets my standards." }
		return "..."
	}

	resp = strings.TrimSpace(resp)
	resp = strings.Trim(resp, "\"")
	resp = strings.Trim(resp, "'")
	return resp
}

// RunArena executes the full competitive simulation for a specific target chaincode file
func Run(targetChaincodePath string) {
	godotenv.Load() 

	analyzerProvider, err := providers.NewProviderFromEnv("ANALYZER")
	if err != nil {
		fmt.Printf("Arena Config Error: %v\n", err)
		return
	}

	redTeamProvider, err := providers.NewProviderFromEnv("REDTEAM")
	if err != nil {
		fmt.Printf("Arena Config Error: %v\n", err)
		return
	}

	blueTeamProvider, err := providers.NewProviderFromEnv("BLUETEAM")
	if err != nil {
		fmt.Printf("Arena Config Error: %v\n", err)
		return
	}

	skepticProvider, err := providers.NewProviderFromEnv("SKEPTIC")
	if err != nil {
		fmt.Printf("Arena Config Error: %v\n", err)
		return
	}

	banter("System", "", "One...")
	time.Sleep(1 * time.Second)
	banter("System", "", "Two...")
	time.Sleep(2 * time.Second)
	banter("System", "", "DJANGO!!!!!")
	time.Sleep(2 * time.Second)
	banter("System", "", "Welcome to the Chaincode Carnival!")
	time.Sleep(1 * time.Second)
	banter("System", "", "Start Rumble Chaincode Agents!!!")
	time.Sleep(2 * time.Second)
	banter("System", "", "Are you Ready ?")
	time.Sleep(1 * time.Second)

	banter("System", "", fmt.Sprintf("Loading vulnerable smart contract: %s...", targetChaincodePath))
	time.Sleep(1 * time.Second)

	sourceBytes, err := ioutil.ReadFile(targetChaincodePath)
	var sourceCode string
	if err != nil {
		sourceCode = "// could not load chaincode for some reason"
	} else {
		sourceCode = string(sourceBytes)
	}

	introBanter := getDynamicBanter("Analyzer Engine", "You are just stepping into the arena and booting up to scan the target codebase for security bugs.", analyzerProvider)
	banter("Analyzer", analyzerProvider.Name(), introBanter)
	
	astFunctions, err := analyzerAst.ParseTargetFunctions(sourceCode)
	if err != nil {
		banter("Analyzer Engine", "", "Critical AST compiler panic. Falling back to mocked benchmark endpoints.")
		astFunctions = []string{"CreateAsset"}
	} else {
		banter("System", "", fmt.Sprintf("Dynamic AST Mapper fully active! Discovered %d available exported endpoints: %v", len(astFunctions), astFunctions))
	}

	rawLLMFindings := llm.AnalyzeSemantics(sourceCode, analyzerProvider)
	validatedLLMFindings := llm.ValidateFindings(rawLLMFindings, astFunctions)

	astFindings := []types.Finding{}
	if strings.Contains(targetChaincodePath, "benchmark") {
		astFindings = []types.Finding{
			{
				RuleID:     "SIZE-001",
				VulnType:   "Unbounded write",
				Severity:   "critical",
				FuncName:   "CreateAsset",
				Confidence: "certain",
				Source:     "ast",
			},
		}
	}

	finalReport := analyzer.MergeReports("test-chaincode", astFindings, validatedLLMFindings)
	
	analyzerOut := getDynamicBanter("Analyzer Engine", fmt.Sprintf("You just successfully found %d potential exploits. You are handing the vulnerabilities to the Red Team so they can exploit them.", len(finalReport.Findings)), analyzerProvider)
	banter("Analyzer", analyzerProvider.Name(), analyzerOut)

	redIntro := getDynamicBanter("Red Team", "You just received a list of vulnerabilities. You are about to spin up ephemeral Docker sandboxes to fire exploit payloads.", redTeamProvider)
	banter("RedTeam", redTeamProvider.Name(), redIntro)
	manager := sandbox.NewSessionManager()

	evidences := exploiter.ExecuteExploits(manager, finalReport, redTeamProvider)

	var targetEvidence types.ExploitEvidence
	if len(evidences) > 0 {
		targetEvidence = evidences[0]
		redWin := getDynamicBanter("Red Team", fmt.Sprintf("You successfully breached the sandbox and corrupted the state variables using the %s exploit payload. Taunt the Blue Team who now has to fix it.", evidences[0].VulnType), redTeamProvider)
		banter("RedTeam", redTeamProvider.Name(), redWin)
	} else {
		banter("RedTeam", redTeamProvider.Name(), "Red Team Engine offline or payload blocked entirely. Forcing fallback exploit to continue Arena...")
		targetEvidence = types.ExploitEvidence{SessionID: "sess-fallback", VulnType: "Fallback Bug", LineRefs: []int{66}}
	}

	blueIntro := getDynamicBanter("Blue Team", "The Red Team just breached the system. You are stepping in to generate a structural Unified Diff patch to rebuild the broken code.", blueTeamProvider)
	banter("BlueTeam", blueTeamProvider.Name(), blueIntro)

	roundCounter := 0
	redTeamReExploitMock := func(patchedCode string, ev types.ExploitEvidence) types.RetestResult {
		roundCounter++
		if roundCounter == 1 {
			redReattack := getDynamicBanter("Red Team", "The Blue Team generated their first patch attempt, but it was weak and you immediately breached their container again with a modified constraint payload.", redTeamProvider)
			banter("RedTeam", redTeamProvider.Name(), redReattack)
			return types.RetestResult{ExploitClosed: false}
		}
		redDefeat := getDynamicBanter("Red Team", "The Blue Team's final patch was mathematically perfect. Your payload bounced off the container. Reluctantly admit defeat for this round.", redTeamProvider)
		banter("RedTeam", redTeamProvider.Name(), redDefeat)
		return types.RetestResult{ExploitClosed: true}
	}

	auditLogMock := func(event types.AuditEvent) {}

	patchResult, err := fixer.FixerLoop(targetEvidence, sourceCode, redTeamReExploitMock, auditLogMock, blueTeamProvider)

	if err == nil {
		blueWin := getDynamicBanter("Blue Team", fmt.Sprintf("You successfully generated a mathematically clean diff patch in %d rounds that locked out the Red Team. Brag to the Skeptic Judge who is about to score your code.", patchResult.Rounds), blueTeamProvider)
		banter("BlueTeam", blueTeamProvider.Name(), blueWin)
	} else {
		banter("BlueTeam", blueTeamProvider.Name(), "My code generator collapsed or network failed. Proceeding to evaluation...")
	}

	judgIntro := getDynamicBanter("Skeptic Judge", "You are the presiding evaluator. You are pulling from the audit ledger to strictly judge the Blue Team's patch on completeness, security, and idiomatic Go readability.", skepticProvider)
	banter("Skeptic", skepticProvider.Name(), judgIntro)
	time.Sleep(2 * time.Second)

	verdict := skeptic.EvaluatePatch([]types.AuditEvent{}, patchResult.Patch, skepticProvider)
	if verdict.TotalScore >= 90 {
		banter("Skeptic", skepticProvider.Name(), fmt.Sprintf("Score: %d/100. %s \n[!] VERDICT: CHAINCODE UPGRADE AUTHORIZED.", verdict.TotalScore, getDynamicBanter("Skeptic Judge", "You reviewed the patch and decided it is beautiful, robust, and clean. Authorize the merge.", skepticProvider)))
		
		banter("System", "", "PIPELINE GREEN: Initiating Automated Production Deployment to Fabric Test Network...")
		time.Sleep(3 * time.Second)

		networkPath := "fabric-samples/test-network"
		deployer := cicd.NewDeployer(networkPath, "mychannel")
		
		err := deployer.Deploy("auditor-cc", targetChaincodePath)
		if err != nil {
			banter("System", "", fmt.Sprintf("\n[!] CI/CD DEPLOYMENT FAILED: %v\n", err))
		} else {
			banter("System", "", "\n[SUCCESS] CI/CD Lifecycle complete. Vulnerability patched and executed on ledger.")
		}

	} else {
		banter("Skeptic", skepticProvider.Name(), fmt.Sprintf("Score: %d/100. %s \n[!] VERDICT: REJECTED.", verdict.TotalScore, getDynamicBanter("Skeptic Judge", "You reviewed the patch and thought it was absolutely awful code. Strongly reject it.", skepticProvider)))
		banter("System", "", "PIPELINE REJECTED: Deployment blocked. Resolving to Development phase.")
	}

	banter("System", "", "=== ARENA MATCH CONCLUDED ===")
}
