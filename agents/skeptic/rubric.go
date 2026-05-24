package skeptic

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/types"
)

type ScoreMetrics struct {
	FixCompleteness     int `json:"fix_completeness"`
	SecurityRegressions int `json:"security_regressions"`
	CodeQuality         int `json:"code_quality"`
	Readability         int `json:"readability"`
}

// EvaluatePatch passes the sandbox evidence to the designated Skeptic LLM model.
func EvaluatePatch(sessionLog []types.AuditEvent, finalPatch string, provider providers.LLMProvider) types.SkepticVerdict {
	prompt := fmt.Sprintf(`You are the presiding Judge over a cybersecurity Red vs Blue match.
The Blue team submitted this unified diff patch:
%s

Score the patch securely on the following criteria (max 100 points total limit). Return only a JSON object matching this schema exactly:
{
  "total_score": 95,
  "feedback": ["Awesome code structure", "Minor variable naming issue"]
}
JSON:`, finalPatch)

	respRaw, err := provider.Query(prompt)
	if err != nil {
		fmt.Printf("[%s] Skeptic query failed: %v\n", provider.Name(), err)
		return types.SkepticVerdict{TotalScore: 0, Approved: false} // Degraded fallback state
	}

	start := strings.Index(respRaw, "{")
	end := strings.LastIndex(respRaw, "}")
	if start == -1 || end == -1 || start > end {
		fmt.Printf("[%s] Skeptic failed to generate recognizable JSON validation.\n", provider.Name())
		return types.SkepticVerdict{TotalScore: 0, Approved: false}
	}
	match := sanitizeSkepticJSON(respRaw[start : end+1])

	var verdict types.SkepticVerdict
	if err := json.Unmarshal([]byte(match), &verdict); err != nil {
		fmt.Printf("[%s] Invalid JSON parsed from Skeptic rubric.\n", provider.Name())
		return types.SkepticVerdict{TotalScore: 0, Approved: false}
	}

	verdict.Approved = verdict.TotalScore >= 90
	
	// Appending legacy checks to handle hallucination structures ensuring test fallback compatibility
	if strings.Contains(finalPatch, "_hack") || strings.Contains(finalPatch, "TODO") {
	    if verdict.TotalScore >= 90 {
	        verdict.TotalScore = 80
	        verdict.Approved = false
	        verdict.Feedback = append(verdict.Feedback, "[OVERRIDE GATE] Identified incomplete/hack resolution constraints in patch.")
	    }
	}

	return verdict
}

func sanitizeSkepticJSON(s string) string {
	if !utf8.ValidString(s) {
		var b strings.Builder
		for _, r := range s {
			if r != utf8.RuneError {
				b.WriteRune(r)
			}
		}
		s = b.String()
	}
	s = fixSkepticEscapes(s)
	trailing := regexp.MustCompile(`,\s*([\]}])`)
	return trailing.ReplaceAllString(s, "$1")
}

func fixSkepticEscapes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inStr := false
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '"' && (i == 0 || s[i-1] != '\\') {
			inStr = !inStr
			b.WriteByte(c)
			i++
			continue
		}
		if inStr && c == '\\' && i+1 < len(s) {
			next := s[i+1]
			switch next {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				b.WriteByte(c); b.WriteByte(next); i += 2
			case 'u':
				if i+5 < len(s) && isSkepticHex(s[i+2]) && isSkepticHex(s[i+3]) && isSkepticHex(s[i+4]) && isSkepticHex(s[i+5]) {
					b.WriteString(s[i : i+6]); i += 6
				} else {
					b.WriteByte(next); i += 2
				}
			default:
				b.WriteByte(next); i += 2
			}
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

func isSkepticHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
