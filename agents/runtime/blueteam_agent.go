package runtime

import (
	"context"
	"fmt"

	"github.com/antonioforte/chaincode-carnival/agents/bus"
	"github.com/antonioforte/chaincode-carnival/agents/fixer"
	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/types"
)

// BlueTeamAgent reacts to FINDINGS_READY (start preparing defense early) and
// EXPLOIT_CONFIRMED (submit actual patch). It runs concurrently with the Red Team.
type BlueTeamAgent struct {
	// retestResults receives RETEST_RESULT from the Red Team
	retestCh chan bus.Event
	// round counter per session
	round int
}

func (a *BlueTeamAgent) Name() string { return "BlueTeam" }

func (a *BlueTeamAgent) Run(ctx context.Context, b *bus.EventBus, provider providers.LLMProvider) {
	findingsCh := make(chan bus.Event, 5)
	exploitCh := make(chan bus.Event, 5)
	retestCh := make(chan bus.Event, 5)

	// Blue Team also subscribes to FINDINGS_READY — it starts working IN PARALLEL with Red Team
	b.Subscribe(bus.EvtFindingsReady, findingsCh)
	b.Subscribe(bus.EvtExploitConfirmed, exploitCh)
	b.Subscribe(bus.EvtRetestResult, retestCh)

	banter := NewBanter("Blue Team", provider)
	fmt.Printf("[BlueTeam] 🛡️  Online. Will begin pre-analysis as soon as findings arrive...\n")

	// Store state across events
	var sourceCode string
	var currentEvidence types.ExploitEvidence

	for {
		select {
		case <-ctx.Done():
			return

		case evt := <-findingsCh:
			// Blue Team reacts to findings in parallel with Red Team —
			// it starts reading the source so it's ready to patch faster.
			var payload FindingsPayload
			if err := remarshal(evt.Payload, &payload); err != nil {
				continue
			}
			sourceCode = payload.SourceCode
			a.round = 0

			b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
				Message: banter.Say(fmt.Sprintf(
					"The Analyzer found %d vulnerabilities — you are already reading the source before the Red Team even launches their first payload.",
					len(payload.Report.Findings),
				)),
			}))

		case evt := <-exploitCh:
			// Red Team confirmed an exploit — now generate the real patch
			var payload ExploitPayload
			if err := remarshal(evt.Payload, &payload); err != nil {
				continue
			}
			currentEvidence = payload.Evidence
			if sourceCode == "" {
				sourceCode = payload.SourceCode
			}

			go a.generateAndSubmitPatch(ctx, sourceCode, currentEvidence, b, provider, banter)

		case evt := <-retestCh:
			// Red Team retested the patch
			var payload RetestPayload
			if err := remarshal(evt.Payload, &payload); err != nil {
				continue
			}

			if payload.ExploitClosed {
				b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
					Message: banter.Say(fmt.Sprintf(
						"Your patch for %s held — the Red Team's payload bounced off the container. Defense complete after %d rounds.",
						payload.Evidence.VulnType, a.round,
					)),
				}))
			} else {
				// Patch failed — generate a new one
				a.round++
				if a.round < 5 {
					b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
						Message: banter.Say(fmt.Sprintf(
							"The Red Team broke through your round %d patch for %s — escalating constraints and regenerating.",
							a.round, payload.Evidence.VulnType,
						)),
					}))
					go a.generateAndSubmitPatch(ctx, sourceCode, payload.Evidence, b, provider, banter)
				} else {
					b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
						Message: banter.Say("Max patch rounds reached — escalating to human review. The Red Team won this round."),
					}))
					// Submit final (possibly imperfect) patch to Skeptic anyway
					b.Publish(bus.NewEvent(bus.EvtPatchSubmitted, a.Name(), PatchPayload{
						Patch:    payload.Patch,
						Rounds:   a.round,
						Evidence: payload.Evidence,
					}))
				}
			}
		}
	}
}

func (a *BlueTeamAgent) generateAndSubmitPatch(
	ctx context.Context,
	sourceCode string,
	evidence types.ExploitEvidence,
	b *bus.EventBus,
	provider providers.LLMProvider,
	banter *BanterService,
) {
	patch, err := fixer.GeneratePatch(evidence, sourceCode, "", a.round, provider)
	if err != nil {
		fmt.Printf("[BlueTeam] Patch generation failed (round %d): %v\n", a.round, err)
		return
	}

	b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
		Message: banter.Say(fmt.Sprintf(
			"You generated a Unified Diff patch targeting the %s vulnerability in round %d — submitting to Red Team for retest.",
			evidence.VulnType, a.round+1,
		)),
	}))

	// Submit to Red Team for retest AND to Skeptic for final evaluation
	b.Publish(bus.NewEvent(bus.EvtPatchSubmitted, a.Name(), PatchPayload{
		Patch:    patch,
		Rounds:   a.round + 1,
		Evidence: evidence,
	}))
}
