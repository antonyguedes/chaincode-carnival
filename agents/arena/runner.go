package arena

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/antonioforte/chaincode-carnival/agents/bus"
	"github.com/antonioforte/chaincode-carnival/agents/dashboard"
	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/agents/runtime"
	"github.com/joho/godotenv"
)

const (
	dashboardPort  = ":8080"
	arenaTimeout   = 10 * time.Minute
	agentBootDelay = 400 * time.Millisecond // wait for all goroutines to be ready
)

// Run launches the full concurrent arena for the given chaincode file.
// All 4 agents start simultaneously and communicate exclusively via the event bus.
func Run(targetChaincodePath string) {
	godotenv.Load()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       ⚔️  CHAINCODE CARNIVAL — CONCURRENT ARENA      ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// ── 1. Build providers from .env ──────────────────────────────────────
	analyzerProvider, err := providers.NewProviderFromEnv("ANALYZER")
	if err != nil { fatal("ANALYZER", err); return }
	redTeamProvider, err := providers.NewProviderFromEnv("REDTEAM")
	if err != nil { fatal("REDTEAM", err); return }
	blueTeamProvider, err := providers.NewProviderFromEnv("BLUETEAM")
	if err != nil { fatal("BLUETEAM", err); return }
	skepticProvider, err := providers.NewProviderFromEnv("SKEPTIC")
	if err != nil { fatal("SKEPTIC", err); return }

	// ── 2. Event Bus ──────────────────────────────────────────────────────
	b := bus.New()

	// ── 3. Dashboard ──────────────────────────────────────────────────────
	hub := dashboard.NewHub()
	hub.ForwardEvents(b)
	go hub.Run()
	go func() {
		fmt.Printf("🖥️  Live dashboard → http://localhost%s\n\n", dashboardPort)
		if err := http.ListenAndServe(dashboardPort, dashboard.Routes(hub)); err != nil {
			fmt.Printf("[Dashboard] Server error: %v\n", err)
		}
	}()
	time.Sleep(200 * time.Millisecond) // give the HTTP server a moment to bind

	// ── 4. Load chaincode source and strip comments ───────────────────────
	// Comments are stripped BEFORE sending to any agent. This prevents the
	// Analyzer from reading hints like "// Vulnerability [MVCC-001]:..." and
	// forces it to discover vulnerabilities through genuine code analysis.
	srcBytes, err := os.ReadFile(targetChaincodePath)
	sourceCode := ""
	if err != nil {
		fmt.Printf("[Arena] Warning: could not read %s: %v\n", targetChaincodePath, err)
		sourceCode = "// file not found"
	} else {
		clean, stripErr := stripGoComments(string(srcBytes))
		if stripErr != nil {
			fmt.Printf("[Arena] Warning: comment stripping failed, using raw source: %v\n", stripErr)
			sourceCode = string(srcBytes)
		} else {
			sourceCode = clean
			fmt.Printf("[Arena] ✂️  Comments stripped from source. Analyzer will work blind.\n")
		}
	}

	// ── 5. Context with global timeout ───────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), arenaTimeout)
	defer cancel()

	// Subscribe to ARENA_OVER to cancel context early when the match ends
	overCh := make(chan bus.Event, 1)
	b.Subscribe(bus.EvtArenaOver, overCh)
	go func() {
		select {
		case <-overCh:
			fmt.Println("\n[Arena] Match concluded. Shutting down agents...")
			cancel()
		case <-ctx.Done():
		}
	}()

	// ── 6. Launch all 4 agents as independent goroutines ─────────────────
	type agentDef struct {
		agent    runtime.Agent
		provider providers.LLMProvider
	}

	agentDefs := []agentDef{
		{&runtime.AnalyzerAgent{}, analyzerProvider},
		{&runtime.RedTeamAgent{}, redTeamProvider},
		{&runtime.BlueTeamAgent{}, blueTeamProvider},
		{runtime.NewSkepticAgent("fabric-samples/test-network", "mychannel"), skepticProvider},
	}

	var wg sync.WaitGroup
	for _, def := range agentDefs {
		wg.Add(1)
		d := def
		go func() {
			defer wg.Done()
			d.agent.Run(ctx, b, d.provider)
		}()
	}

	// Give all goroutines time to subscribe before publishing the start event
	time.Sleep(agentBootDelay)

	// ── 7. Fire the starting pistol ───────────────────────────────────────
	fmt.Printf("[Arena] 🔫 All agents online. Firing ARENA_START for: %s\n\n", targetChaincodePath)
	b.Publish(bus.NewEvent(bus.EvtArenaStart, "System", runtime.ArenaStartPayload{
		ChaincodePath: targetChaincodePath,
		SourceCode:    sourceCode,
	}))

	// ── 8. Block until all agents finish ─────────────────────────────────
	wg.Wait()
	fmt.Println("\n[Arena] === MATCH OVER ===")
}

func fatal(prefix string, err error) {
	fmt.Printf("[Arena] Config error for %s: %v\n", strings.ToUpper(prefix), err)
}

// stripGoComments uses the official Go parser to remove ALL comments from
// source code — both // line comments and /* block comments */.
// This prevents the Analyzer from reading hints like "// Vulnerability [X]:"
// embedded in the chaincode, forcing genuine blind discovery.
func stripGoComments(src string) (string, error) {
	fset := token.NewFileSet()
	// NOT passing parser.ParseComments — the AST will have no comment nodes
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return src, fmt.Errorf("parser failed: %w", err)
	}
	// Clear any comment groups that snuck in
	f.Comments = []*ast.CommentGroup{}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, f); err != nil {
		return src, fmt.Errorf("printer failed: %w", err)
	}
	return buf.String(), nil
}
