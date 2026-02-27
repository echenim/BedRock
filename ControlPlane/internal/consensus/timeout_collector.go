package consensus

import (
	"fmt"

	"github.com/echenim/Bedrock/controlplane/internal/types"
)

// TimeoutCollector collects timeout messages for a specific (height, round) and
// determines when the f+1 threshold is reached, forming a Timeout Certificate (TC).
// Per SPEC.md §10: f+1 timeout messages are required before any honest node advances
// to a new round. This prevents a single Byzantine validator from forcing round skips.
type TimeoutCollector struct {
	height       uint64
	round        uint64
	valSet       *types.ValidatorSet
	timeouts     map[types.Address]*types.TimeoutMessage
	timeoutPower uint64
}

// NewTimeoutCollector creates a new TimeoutCollector for the given height and round.
func NewTimeoutCollector(height, round uint64, valSet *types.ValidatorSet) *TimeoutCollector {
	return &TimeoutCollector{
		height:   height,
		round:    round,
		valSet:   valSet,
		timeouts: make(map[types.Address]*types.TimeoutMessage),
	}
}

// AddTimeout adds a timeout message to the collector.
// Returns (thresholdReached, error). The threshold is f+1 voting power.
func (tc *TimeoutCollector) AddTimeout(msg *types.TimeoutMessage) (bool, error) {
	if msg.Height != tc.height || msg.Round != tc.round {
		return false, fmt.Errorf("timeout for (h=%d, r=%d) does not match collector (h=%d, r=%d)",
			msg.Height, msg.Round, tc.height, tc.round)
	}

	// Look up validator.
	val, ok := tc.valSet.GetByAddress(msg.VoterID)
	if !ok {
		return false, fmt.Errorf("timeout from unknown validator %s", msg.VoterID)
	}

	// Duplicate check — ignore repeated messages from the same validator.
	if _, exists := tc.timeouts[msg.VoterID]; exists {
		return tc.HasThreshold(), nil
	}

	tc.timeouts[msg.VoterID] = msg
	tc.timeoutPower += val.VotingPower

	return tc.HasThreshold(), nil
}

// HasThreshold returns true if collected timeouts have >= f+1 power.
func (tc *TimeoutCollector) HasThreshold() bool {
	return tc.valSet.HasFPlusOne(tc.timeoutPower)
}

// HighestQC returns the highest *verified* QC carried by any collected timeout
// message. QCs that fail signature/quorum verification against the validator set
// are silently skipped to prevent fake QC injection (audit S4).
func (tc *TimeoutCollector) HighestQC() *types.QuorumCertificate {
	var best *types.QuorumCertificate
	for _, msg := range tc.timeouts {
		if msg.HighQC == nil {
			continue
		}
		// Verify the QC before trusting it.
		if err := msg.HighQC.Verify(tc.valSet); err != nil {
			continue
		}
		if best == nil || msg.HighQC.Round > best.Round {
			best = msg.HighQC
		}
	}
	return best
}

// Size returns the number of timeout messages collected.
func (tc *TimeoutCollector) Size() int {
	return len(tc.timeouts)
}
