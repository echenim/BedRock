package p2p

import (
	"testing"
)

// FuzzDecodeMessage fuzzes the top-level P2P message decoder.
// This exercises DecodeEnvelope → per-type decoders (proposal, vote, timeout).
func FuzzDecodeMessage(f *testing.F) {
	// Seed corpus: valid envelope prefix bytes for each message type.
	f.Add([]byte{byte(MsgProposal), 0x0a, 0x00})
	f.Add([]byte{byte(MsgVote), 0x0a, 0x00})
	f.Add([]byte{byte(MsgTimeout), 0x0a, 0x00})
	// Empty and minimal inputs.
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add([]byte{0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic on any input.
		_, _, _ = DecodeMessage(data)
	})
}

// FuzzDecodeEnvelope fuzzes the envelope decoder with per-type size limits.
func FuzzDecodeEnvelope(f *testing.F) {
	f.Add([]byte{byte(MsgProposal), 0x08, 0x01})
	f.Add([]byte{byte(MsgVote), 0x08, 0x01})
	f.Add([]byte{byte(MsgTimeout), 0x08, 0x01})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeEnvelope(data)
	})
}

// FuzzDecodeProposal fuzzes the protobuf proposal decoder.
func FuzzDecodeProposal(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x0a, 0x00})
	f.Add([]byte{0x08, 0x01, 0x10, 0x02})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeProposal(data)
	})
}

// FuzzDecodeVote fuzzes the protobuf vote decoder.
func FuzzDecodeVote(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x0a, 0x00})
	f.Add([]byte{0x08, 0x01, 0x10, 0x02})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeVote(data)
	})
}

// FuzzDecodeTimeout fuzzes the protobuf timeout message decoder.
func FuzzDecodeTimeout(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x0a, 0x00})
	f.Add([]byte{0x08, 0x01, 0x10, 0x02})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeTimeout(data)
	})
}
