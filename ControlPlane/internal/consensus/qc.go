package consensus

import (
	"errors"

	"github.com/echenim/Bedrock/controlplane/internal/types"
)

// MakeQC creates a QuorumCertificate from the collected votes.
// Only valid if the VoteSet HasQuorum().
// Builds the vote slice directly from the map to avoid an intermediate
// pointer slice allocation (audit P3).
func (vs *VoteSet) MakeQC() (*types.QuorumCertificate, error) {
	if !vs.HasQuorum() {
		return nil, errors.New("cannot create QC: insufficient quorum")
	}

	n := len(vs.votes)
	if n == 0 {
		return nil, errors.New("cannot create QC: no votes")
	}

	domainVotes := make([]types.Vote, 0, n)
	var blockHash types.Hash
	first := true
	for _, v := range vs.votes {
		if first {
			blockHash = v.BlockHash
			first = false
		}
		domainVotes = append(domainVotes, *v)
	}

	return &types.QuorumCertificate{
		BlockHash: blockHash,
		Round:     vs.round,
		Votes:     domainVotes,
	}, nil
}

// ForkChoice selects the preferred chain based on QC height.
// Per SPEC-v0.2.md §9: highest QC height, then highest round, then hash.
func ForkChoice(a, b *types.QuorumCertificate) *types.QuorumCertificate {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	// Prefer higher round QC.
	if a.Round > b.Round {
		return a
	}
	if b.Round > a.Round {
		return b
	}

	// Tie-break by block hash (deterministic).
	for i := range a.BlockHash {
		if a.BlockHash[i] > b.BlockHash[i] {
			return a
		}
		if b.BlockHash[i] > a.BlockHash[i] {
			return b
		}
	}
	return a
}
