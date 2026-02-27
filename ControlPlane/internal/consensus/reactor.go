package consensus

import (
	"context"

	"github.com/echenim/Bedrock/controlplane/internal/types"
	"go.uber.org/zap"
)

// HandleProposal processes a received proposal message.
func (e *Engine) HandleProposal(proposal *types.Proposal) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if proposal == nil || proposal.Block == nil {
		return
	}

	// Ignore proposals for wrong height/round.
	if proposal.Block.Header.Height != e.state.Height {
		e.logger.Debug("ignoring proposal for wrong height",
			zap.Uint64("got", proposal.Block.Header.Height),
			zap.Uint64("want", e.state.Height),
		)
		return
	}
	if proposal.Round != e.state.Round {
		e.logger.Debug("ignoring proposal for wrong round",
			zap.Uint64("got", proposal.Round),
			zap.Uint64("want", e.state.Round),
		)
		return
	}

	// Already have a proposal for this round.
	if e.state.Proposal != nil {
		return
	}

	// Validate proposal.
	if err := e.ValidateProposal(proposal); err != nil {
		e.logger.Warn("invalid proposal", zap.Error(err))
		return
	}

	e.state.Proposal = proposal

	// If we got the proposal and were waiting, move to vote.
	if e.state.Step == StepPropose {
		e.EnterVote()
	}
}

// HandleVote processes a received vote message.
func (e *Engine) HandleVote(vote *types.Vote) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if vote == nil {
		return
	}

	// Ignore votes for wrong height/round.
	if vote.Height != e.state.Height || vote.Round != e.state.Round {
		return
	}

	quorum, evidence, err := e.state.VoteSet.AddVote(vote)
	if err != nil {
		e.logger.Debug("failed to add vote", zap.Error(err))
		return
	}

	if evidence != nil {
		e.logger.Warn("equivocation detected",
			zap.String("validator", vote.VoterID.String()),
		)
		e.evidencePool.AddEvidence(evidence)
	}

	if quorum && e.state.Step == StepVote {
		e.onQuorumReached()
	}
}

// HandleTimeoutMsg processes a received timeout message from a peer.
// Per SPEC.md §10: f+1 timeout messages for the current round are required to
// form a Timeout Certificate (TC) before advancing to the next round. This prevents
// a single Byzantine validator from forcing arbitrary round skips.
func (e *Engine) HandleTimeoutMsg(msg *types.TimeoutMessage) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if msg == nil {
		return
	}

	// Only process timeouts for our current height and round.
	if msg.Height != e.state.Height || msg.Round != e.state.Round {
		e.logger.Debug("ignoring timeout for different height/round",
			zap.Uint64("msg_height", msg.Height),
			zap.Uint64("msg_round", msg.Round),
			zap.Uint64("our_height", e.state.Height),
			zap.Uint64("our_round", e.state.Round),
		)
		return
	}

	// Add to timeout collector for this round.
	threshold, err := e.state.TimeoutCollector.AddTimeout(msg)
	if err != nil {
		e.logger.Debug("failed to add timeout message", zap.Error(err))
		return
	}

	if !threshold {
		e.logger.Debug("timeout message collected, waiting for f+1 threshold",
			zap.Int("collected", e.state.TimeoutCollector.Size()),
			zap.Uint64("height", msg.Height),
			zap.Uint64("round", msg.Round),
		)
		return
	}

	// f+1 threshold reached — form TC and advance.
	e.logger.Info("f+1 timeout threshold reached, advancing round",
		zap.Uint64("height", e.state.Height),
		zap.Uint64("from_round", e.state.Round),
		zap.Uint64("to_round", e.state.Round+1),
		zap.Int("timeout_count", e.state.TimeoutCollector.Size()),
	)

	// Update highest QC from the TC's collected timeouts.
	if highQC := e.state.TimeoutCollector.HighestQC(); highQC != nil {
		e.state.UpdateHighestQC(highQC)
	}

	e.EnterNewRound(e.state.Round + 1)
}

// eventLoop is the main consensus event loop.
// All state mutations happen through this goroutine to prevent races.
func (e *Engine) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case proposal := <-e.proposalCh:
			e.HandleProposal(proposal)

		case vote := <-e.voteCh:
			e.HandleVote(vote)

		case te := <-e.timeoutCh:
			e.mu.Lock()
			e.HandleTimeout(te.Height, te.Round)
			e.mu.Unlock()

		case <-e.nextHeightCh:
			e.mu.Lock()
			e.EnterPropose()
			e.mu.Unlock()
		}
	}
}

// DrainNextHeight processes a pending next-height signal synchronously.
// Used in tests to step through the two-chain commit rule.
func (e *Engine) DrainNextHeight() bool {
	select {
	case <-e.nextHeightCh:
		e.EnterPropose()
		return true
	default:
		return false
	}
}

// SubmitProposal queues a proposal for processing.
func (e *Engine) SubmitProposal(proposal *types.Proposal) {
	select {
	case e.proposalCh <- proposal:
	default:
		e.logger.Warn("proposal channel full, dropping")
	}
}

// SubmitVote queues a vote for processing.
func (e *Engine) SubmitVote(vote *types.Vote) {
	select {
	case e.voteCh <- vote:
	default:
		e.logger.Warn("vote channel full, dropping")
	}
}
