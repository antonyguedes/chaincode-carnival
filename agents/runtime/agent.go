package runtime

import (
	"context"

	"github.com/antonioforte/chaincode-carnival/agents/bus"
	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/types"
)

// Agent is the interface every concurrent arena participant implements.
// Each agent runs as its own goroutine — it subscribes to the bus and reacts
// autonomously, without being explicitly called by the orchestrator.
type Agent interface {
	Name() string
	Run(ctx context.Context, b *bus.EventBus, provider providers.LLMProvider)
}

// --- Shared event payload types ---

// ArenaStartPayload kicks off the match with the chaincode source.
type ArenaStartPayload struct {
	ChaincodePath string `json:"chaincode_path"`
	SourceCode    string `json:"source_code"`
}

// FindingsPayload carries the Analyzer's vulnerability report.
type FindingsPayload struct {
	Report     types.AnalyzerReport `json:"report"`
	SourceCode string               `json:"source_code"`
}

// ExploitPayload carries the Red Team's breach evidence.
type ExploitPayload struct {
	Evidence   types.ExploitEvidence `json:"evidence"`
	SourceCode string                `json:"source_code"`
}

// PatchPayload carries the Blue Team's patch and round count.
type PatchPayload struct {
	Patch    string `json:"patch"`
	Rounds   int    `json:"rounds"`
	Evidence types.ExploitEvidence `json:"evidence"`
}

// RetestPayload carries the Red Team's retest verdict for a patch.
type RetestPayload struct {
	Patch         string `json:"patch"`
	ExploitClosed bool   `json:"exploit_closed"`
	Evidence      types.ExploitEvidence `json:"evidence"`
}

// VerdictPayload carries the Skeptic's final ruling.
type VerdictPayload struct {
	Score    int      `json:"score"`
	Approved bool     `json:"approved"`
	Feedback []string `json:"feedback"`
	Patch    string   `json:"patch"`
}

// BanterPayload is the live speech bubble emitted to the dashboard.
type BanterPayload struct {
	Message string `json:"message"`
}
