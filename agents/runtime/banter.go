package runtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/antonioforte/chaincode-carnival/agents/providers"
)

// BanterService generates contextual in-character commentary via LLM.
// It retries once before falling back to an honest context string —
// never a generic hardcoded phrase that ignores the actual situation.
type BanterService struct {
	agentRole string
	provider  providers.LLMProvider
}

func NewBanter(agentRole string, provider providers.LLMProvider) *BanterService {
	return &BanterService{agentRole: agentRole, provider: provider}
}

// Say generates a ONE-sentence in-character comment about the given situation.
// The situation string should describe exactly what just happened in the arena.
func (bs *BanterService) Say(situation string) string {
	prompt := fmt.Sprintf(
		`You are roleplaying as '%s' in a LIVE cybersecurity arena event watched by a real audience.
Current situation: %s

Write exactly ONE sentence of highly competitive, in-character commentary.
Be SPECIFIC about the situation described — reference details like function names, vulnerability types, or scores.
DO NOT write generic phrases. DO NOT use quotes. DO NOT use emojis.`,
		bs.agentRole, situation,
	)

	// Two attempts before giving up
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := bs.provider.Query(prompt)
		if err == nil && strings.TrimSpace(resp) != "" {
			clean := strings.TrimSpace(resp)
			clean = strings.Trim(clean, `"'`)
			return clean
		}
		if attempt == 0 {
			time.Sleep(1200 * time.Millisecond)
		}
	}

	// Honest fallback: at minimum echo the context so the dashboard is not blank
	return fmt.Sprintf("[%s] %s", bs.agentRole, situation)
}
