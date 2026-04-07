package types

import "time"

type AnalyzerReport struct {
	ChaincodeID  string    `json:"chaincode_id"`
	Timestamp    time.Time `json:"timestamp"`
	Findings     []Finding `json:"findings"`
	RiskScore    int       `json:"risk_score"`    // 0–100
	ExploitOrder []string  `json:"exploit_order"` // rule IDs, priority order
}

type Finding struct {
	RuleID          string `json:"rule_id"`
	VulnType        string `json:"vuln_type"`
	Severity        string `json:"severity"`   // critical | high | medium
	LineRefs        []int  `json:"line_refs"`
	FuncName        string `json:"func_name"`
	Snippet         string `json:"snippet"`
	Confidence      string `json:"confidence"` // certain | likely | possible
	Source          string `json:"source"`     // ast | llm_semantic
	NeedsConcurrent bool   `json:"needs_concurrent"`
}

type ExploitEvidence struct {
	SessionID       string `json:"session_id"`
	VulnType        string `json:"vuln_type"`
	LineRefs        []int  `json:"line_refs"`
	TxPayload       string `json:"tx_payload"`
	StateDiffBefore string `json:"state_diff_before"`
	StateDiffAfter  string `json:"state_diff_after"`
	PanicTrace      string `json:"panic_trace"`
	Confirmed       bool   `json:"confirmed"`
}

// RetestResult is derived from its usage in the Phase 3 implementation logic
type RetestResult struct {
	ExploitClosed bool              `json:"exploit_closed"`
	NewVulns      []ExploitEvidence `json:"new_vulns,omitempty"`
}

type BattleRound struct {
	RoundNum       int             `json:"round_num"`
	ExploitAttempt ExploitEvidence `json:"exploit_attempt"`
	PatchDiff      string          `json:"patch_diff"`
	RetestResult   RetestResult    `json:"retest_result"`
}

type SessionRecord struct {
	SessionID         string         `json:"session_id"`
	OriginalSrc       string         `json:"original_src"`
	AnalyzerReport    AnalyzerReport `json:"analyzer_report"`
	Rounds            []BattleRound  `json:"rounds"`
	FinalPatch        string         `json:"final_patch"`
	FinalRetestResult RetestResult   `json:"final_retest_result"`
}

type SkepticVerdict struct {
	TotalScore            int                       `json:"total_score"`
	Approved              bool                      `json:"approved"`
	Verdict               string                    `json:"verdict"` // APPROVE | REJECT | APPROVE_WITH_FLAG | ESCALATE
	Feedback              []string                  `json:"feedback"`
	DimensionScores       map[string]DimensionScore `json:"dimension_scores"`
	RejectionInstructions string                    `json:"rejection_instructions"`
	Flags                 []string                  `json:"flags"`
}

type DimensionScore struct {
	Pass           bool     `json:"pass"`
	Reasoning      string   `json:"reasoning"`
	FailedCriteria []string `json:"failed_criteria"`
}

type AuditEvent struct {
	EventType string    `json:"event_type"` // EXPLOIT_ATTEMPT | PATCH_ROUND | SKEPTIC_RULING | UPGRADE_PROPOSED | HUMAN_REVIEW_REQUIRED
	SessionID string    `json:"session_id"`
	TxID      string    `json:"tx_id"`      // The underlying Fabric transaction ID
	Timestamp time.Time `json:"timestamp"`
	AgentID   string    `json:"agent_id"`
	Payload   string    `json:"payload"` // JSON-encoded event-specific data
}
