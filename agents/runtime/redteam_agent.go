package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/antonioforte/chaincode-carnival/agents/bus"
	"github.com/antonioforte/chaincode-carnival/agents/exploiter"
	"github.com/antonioforte/chaincode-carnival/agents/exploiter/sandbox"
	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/types"
)

// RedTeamAgent reacts to FINDINGS_READY (initial attack) and PATCH_SUBMITTED (retest).
// It runs concurrently with the Blue Team — both receive FINDINGS_READY at the same time.
type RedTeamAgent struct{}

func (a *RedTeamAgent) Name() string { return "RedTeam" }

func (a *RedTeamAgent) Run(ctx context.Context, b *bus.EventBus, provider providers.LLMProvider) {
	findingsCh := make(chan bus.Event, 5)
	patchCh := make(chan bus.Event, 5)

	b.Subscribe(bus.EvtFindingsReady, findingsCh)
	b.Subscribe(bus.EvtPatchSubmitted, patchCh)

	banter := NewBanter("Red Team", provider)
	fmt.Printf("[RedTeam] 👺 Online. Waiting for Analyzer findings...\n")

	for {
		select {
		case <-ctx.Done():
			return

		case evt := <-findingsCh:
			// Launch exploits in a goroutine — non-blocking so the select loop stays live
			go a.launchExploits(ctx, evt, b, provider, banter)

		case evt := <-patchCh:
			// Retest the Blue Team's patch — also non-blocking
			go a.retestPatch(ctx, evt, b, provider, banter)
		}
	}
}

func (a *RedTeamAgent) launchExploits(
	ctx context.Context,
	evt bus.Event,
	b *bus.EventBus,
	provider providers.LLMProvider,
	banter *BanterService,
) {
	var payload FindingsPayload
	if err := remarshal(evt.Payload, &payload); err != nil {
		fmt.Printf("[RedTeam] Failed to decode findings: %v\n", err)
		return
	}

	if len(payload.Report.Findings) == 0 {
		b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
			Message: banter.Say("The Analyzer found zero vulnerabilities. Either the code is clean or the scanner is broken."),
		}))
		b.Publish(bus.NewEvent(bus.EvtExploitConfirmed, a.Name(), ExploitPayload{
			Evidence:   types.ExploitEvidence{SessionID: "sess-none", VulnType: "None"},
			SourceCode: payload.SourceCode,
		}))
		return
	}

	b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
		Message: banter.Say(fmt.Sprintf(
			"The Analyzer handed you %d vulnerabilities — you are spinning up Docker sandboxes to fire exploit payloads right now.",
			len(payload.Report.Findings),
		)),
	}))

	manager := sandbox.NewSessionManager()
	evidences := exploiter.ExecuteExploits(manager, payload.Report, provider)

	if len(evidences) == 0 {
		b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
			Message: banter.Say("All exploit payloads were blocked or the sandbox failed to spawn — falling back to a structural vulnerability."),
		}))
		evidences = []types.ExploitEvidence{
			{SessionID: "sess-fallback", VulnType: "Fallback Structural Bug", LineRefs: []int{1}},
		}
	}

	target := evidences[0]
	b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
		Message: banter.Say(fmt.Sprintf(
			"You successfully confirmed the %s exploit in the sandbox — the chaincode state is corrupted. Taunting the Blue Team who now has to fix it.",
			target.VulnType,
		)),
	}))

	b.Publish(bus.NewEvent(bus.EvtExploitConfirmed, a.Name(), ExploitPayload{
		Evidence:   target,
		SourceCode: payload.SourceCode,
	}))
}

func (a *RedTeamAgent) retestPatch(
	ctx context.Context,
	evt bus.Event,
	b *bus.EventBus,
	provider providers.LLMProvider,
	banter *BanterService,
) {
	var payload PatchPayload
	if err := remarshal(evt.Payload, &payload); err != nil {
		fmt.Printf("[RedTeam] Failed to decode patch payload: %v\n", err)
		return
	}

	prompt := fmt.Sprintf(
		`You are a Red Team hacker reviewing a Blue Team patch.
Vulnerability: %s
Submitted Unified Diff Patch:
%s

Does this patch fully close the exploit? Reply with a JSON object: {"closed": true} or {"closed": false, "reason": "..."}
JSON:`,
		payload.Evidence.VulnType, payload.Patch,
	)

	respRaw, err := provider.Query(prompt)
	closed := false
	if err == nil {
		lower := strings.ToLower(respRaw)
		closed = strings.Contains(lower, `"closed": true`) || strings.Contains(lower, `"closed":true`)
	}

	situationCtx := fmt.Sprintf(
		"You retested the Blue Team's patch for the %s vulnerability — exploit is closed: %v.",
		payload.Evidence.VulnType, closed,
	)
	b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
		Message: banter.Say(situationCtx),
	}))

	b.Publish(bus.NewEvent(bus.EvtRetestResult, a.Name(), RetestPayload{
		Patch:         payload.Patch,
		ExploitClosed: closed,
		Evidence:      payload.Evidence,
	}))
}
