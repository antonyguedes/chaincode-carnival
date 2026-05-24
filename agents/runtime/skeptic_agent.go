package runtime

import (
	"context"
	"fmt"

	"github.com/antonioforte/chaincode-carnival/agents/bus"
	"github.com/antonioforte/chaincode-carnival/agents/providers"
	"github.com/antonioforte/chaincode-carnival/agents/skeptic"
	"github.com/antonioforte/chaincode-carnival/agents/cicd"
)

// SkepticAgent watches every event and delivers a live verdict when a patch is submitted.
// It subscribes to ALL events so it can comment in real-time throughout the match.
type SkepticAgent struct {
	networkPath string
	channelName string
}

func NewSkepticAgent(networkPath, channelName string) *SkepticAgent {
	return &SkepticAgent{networkPath: networkPath, channelName: channelName}
}

func (a *SkepticAgent) Name() string { return "Skeptic" }

func (a *SkepticAgent) Run(ctx context.Context, b *bus.EventBus, provider providers.LLMProvider) {
	patchCh    := make(chan bus.Event, 5)
	exploitCh  := make(chan bus.Event, 5)
	retestCh   := make(chan bus.Event, 5)

	b.Subscribe(bus.EvtPatchSubmitted, patchCh)
	b.Subscribe(bus.EvtExploitConfirmed, exploitCh)
	b.Subscribe(bus.EvtRetestResult, retestCh)

	banter := NewBanter("Skeptic Judge", provider)
	fmt.Printf("[Skeptic] ⚖️  Online. Watching the match...\n")

	for {
		select {
		case <-ctx.Done():
			return

		case evt := <-exploitCh:
			// Live commentary as the exploit lands
			var payload ExploitPayload
			if err := remarshal(evt.Payload, &payload); err != nil {
				continue
			}
			b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
				Message: banter.Say(fmt.Sprintf(
					"Red Team just confirmed the %s exploit in the sandbox — now I watch to see if Blue Team can seal it.",
					payload.Evidence.VulnType,
				)),
			}))

		case evt := <-retestCh:
			// Commentary on the retest result
			var payload RetestPayload
			if err := remarshal(evt.Payload, &payload); err != nil {
				continue
			}
			if payload.ExploitClosed {
				b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
					Message: banter.Say(fmt.Sprintf(
						"Blue Team closed the %s exploit — now I need to formally score the quality of that patch.",
						payload.Evidence.VulnType,
					)),
				}))
			} else {
				b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
					Message: banter.Say(fmt.Sprintf(
						"Red Team broke through the patch for %s again — the Blue Team's defense is leaking.",
						payload.Evidence.VulnType,
					)),
				}))
			}

		case evt := <-patchCh:
			// Only evaluate when the Blue Team explicitly closes the loop
			var payload PatchPayload
			if err := remarshal(evt.Payload, &payload); err != nil {
				continue
			}

			go a.evaluateAndDeliver(ctx, payload, b, provider, banter)
		}
	}
}

func (a *SkepticAgent) evaluateAndDeliver(
	ctx context.Context,
	payload PatchPayload,
	b *bus.EventBus,
	provider providers.LLMProvider,
	banter *BanterService,
) {
	b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{
		Message: banter.Say(fmt.Sprintf(
			"Blue Team submitted their patch for %s after %d rounds — you are scoring it on security, readability, Go idioms, and completeness.",
			payload.Evidence.VulnType, payload.Rounds,
		)),
	}))

	verdict := skeptic.EvaluatePatch(nil, payload.Patch, provider)

	var verdictBanter string
	if verdict.TotalScore >= 90 {
		verdictBanter = banter.Say(fmt.Sprintf(
			"Score %d/100 — the patch for %s is structurally clean and the exploit surface is sealed. Authorizing production deployment.",
			verdict.TotalScore, payload.Evidence.VulnType,
		))
	} else {
		verdictBanter = banter.Say(fmt.Sprintf(
			"Score %d/100 — the patch for %s has unacceptable gaps. Rejecting and sending back to development.",
			verdict.TotalScore, payload.Evidence.VulnType,
		))
	}

	b.Publish(bus.NewEvent(bus.EvtBanter, a.Name(), BanterPayload{Message: verdictBanter}))
	b.Publish(bus.NewEvent(bus.EvtVerdictRendered, a.Name(), VerdictPayload{
		Score:    verdict.TotalScore,
		Approved: verdict.Approved,
		Feedback: verdict.Feedback,
		Patch:    payload.Patch,
	}))

	if verdict.Approved {
		fmt.Printf("\n[CI/CD] 🚀 Score ≥ 90 — initiating Hyperledger Fabric deployment...\n")
		deployer := cicd.NewDeployer(a.networkPath, a.channelName)
		if err := deployer.Deploy("auditor-cc", "chaincodes/hardcore/main.go"); err != nil {
			fmt.Printf("[CI/CD] ❌ Deployment failed: %v\n", err)
		} else {
			fmt.Printf("[CI/CD] ✅ Chaincode deployed to channel '%s'\n", a.channelName)
		}
	}

	b.Publish(bus.NewEvent(bus.EvtArenaOver, a.Name(), nil))
}
