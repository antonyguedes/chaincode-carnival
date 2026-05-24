package runtime

import (
	"context"
	"fmt"

	analyzerPkg "github.com/antonioforte/chaincode-carnival/agents/analyzer"
	analyzerAst "github.com/antonioforte/chaincode-carnival/agents/analyzer/ast"
	analyzerLlm "github.com/antonioforte/chaincode-carnival/agents/analyzer/llm"
	"github.com/antonioforte/chaincode-carnival/agents/bus"
	"github.com/antonioforte/chaincode-carnival/agents/providers"
)

// AnalyzerAgent listens for ARENA_START and publishes FINDINGS_READY.
// It runs the full AST + LLM semantic scan on the chaincode source.
type AnalyzerAgent struct{}

func (a *AnalyzerAgent) Name() string { return "Analyzer" }

func (a *AnalyzerAgent) Run(ctx context.Context, b *bus.EventBus, provider providers.LLMProvider) {
	startCh := make(chan bus.Event, 1)
	b.Subscribe(bus.EvtArenaStart, startCh)

	banter := NewBanter("Analyzer Engine", provider)
	fmt.Printf("[Analyzer] 🔍 Online. Waiting for arena start...\n")

	select {
	case <-ctx.Done():
		return
	case evt := <-startCh:
		var payload ArenaStartPayload
		if err := remarshal(evt.Payload, &payload); err != nil {
			fmt.Printf("[Analyzer] Failed to decode start payload: %v\n", err)
			return
		}

		// Announce to dashboard
		b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
			Message: banter.Say(fmt.Sprintf("You are booting up to scan the target chaincode at %s for critical security vulnerabilities.", payload.ChaincodePath)),
		}))

		// --- Real work: AST parse + LLM semantic analysis ---
		astFunctions, err := analyzerAst.ParseTargetFunctions(payload.SourceCode)
		if err != nil {
			fmt.Printf("[Analyzer] AST parse failed, continuing with empty function list\n")
			astFunctions = []string{}
		} else {
			fmt.Printf("[Analyzer] AST mapped %d exported endpoints: %v\n", len(astFunctions), astFunctions)
		}

		rawFindings := analyzerLlm.AnalyzeSemantics(payload.SourceCode, provider)
		validated := analyzerLlm.ValidateFindings(rawFindings, astFunctions)
		report := analyzerPkg.MergeReports("arena-target", nil, validated)

		situationCtx := fmt.Sprintf(
			"You just finished scanning the chaincode and found %d vulnerabilities. You are handing off to the Red Team.",
			len(report.Findings),
		)
		b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
			Message: banter.Say(situationCtx),
		}))

		// Publish findings — Red Team and Blue Team will react concurrently
		b.Publish(bus.NewEvent(bus.EvtFindingsReady, a.Name(), FindingsPayload{
			Report:     report,
			SourceCode: payload.SourceCode,
		}))
	}
}
