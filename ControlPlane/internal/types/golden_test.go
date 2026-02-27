package types

import (
	"encoding/hex"
	"testing"

	"google.golang.org/protobuf/proto"
)

// Golden test vectors ensure serialization format stability across builds.
// If any of these fail, the serialization format has changed and all nodes
// must be upgraded together (a breaking consensus change).
//
// To regenerate vectors after an intentional format change:
//   go test -run TestGolden -v -update-golden
//
// These vectors encode minimal but representative structures.

func mustHashFromHex(t *testing.T, s string) Hash {
	t.Helper()
	h, err := HashFromHex(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestGoldenBlockHeader(t *testing.T) {
	h := &BlockHeader{
		Height:     42,
		Round:      1,
		ParentHash: mustHashFromHex(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		StateRoot:  mustHashFromHex(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		TxRoot:     mustHashFromHex(t, "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		BlockTime:  1700000000,
		ChainID:    []byte("bedrock-test-1"),
	}

	hash, err := h.ComputeHash()
	if err != nil {
		t.Fatal(err)
	}

	// The block hash must be deterministic across all builds.
	got := hex.EncodeToString(hash[:])
	// Golden value — regenerate if proto schema changes intentionally.
	want := golden(t, "block_header_hash", got)
	if got != want {
		t.Fatalf("block header hash mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestGoldenVoteSerialization(t *testing.T) {
	v := &Vote{
		BlockHash: mustHashFromHex(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Height:    10,
		Round:     2,
	}
	pb := v.ToProto()
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(pb)
	if err != nil {
		t.Fatal(err)
	}

	got := hex.EncodeToString(data)
	want := golden(t, "vote_serialized", got)
	if got != want {
		t.Fatalf("vote serialization mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestGoldenQCSerialization(t *testing.T) {
	qc := &QuorumCertificate{
		BlockHash: mustHashFromHex(t, "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
		Round:     5,
		Votes: []Vote{
			{
				BlockHash: mustHashFromHex(t, "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
				Height:    7,
				Round:     5,
			},
		},
	}
	pb := qc.ToProto()
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(pb)
	if err != nil {
		t.Fatal(err)
	}

	got := hex.EncodeToString(data)
	want := golden(t, "qc_serialized", got)
	if got != want {
		t.Fatalf("QC serialization mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestGoldenProposalSerialization(t *testing.T) {
	p := &Proposal{
		Block: &Block{
			Header: BlockHeader{
				Height:  100,
				Round:   0,
				ChainID: []byte("bedrock-test-1"),
			},
			Transactions: [][]byte{
				[]byte("tx1"),
				[]byte("tx2"),
			},
		},
		Round: 0,
	}
	pb := p.ToProto()
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(pb)
	if err != nil {
		t.Fatal(err)
	}

	got := hex.EncodeToString(data)
	want := golden(t, "proposal_serialized", got)
	if got != want {
		t.Fatalf("proposal serialization mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestGoldenTimeoutSerialization(t *testing.T) {
	tm := &TimeoutMessage{
		Height: 50,
		Round:  3,
	}
	pb := tm.ToProto()
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(pb)
	if err != nil {
		t.Fatal(err)
	}

	got := hex.EncodeToString(data)
	want := golden(t, "timeout_serialized", got)
	if got != want {
		t.Fatalf("timeout serialization mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

// goldenVectors stores the expected hex-encoded serialization for each test.
// These are populated on first run and must be committed to the repository.
var goldenVectors = map[string]string{
	"block_header_hash":    "559f2e775e17d2f746da7cded44fab281c195d79ebf142e9a22288899e836c24",
	"vote_serialized":      "0a20aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa100a1802222000000000000000000000000000000000000000000000000000000000000000002a4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	"qc_serialized":        "0a20dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd10051a640a200000000000000000000000000000000000000000000000000000000000000000124000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	"proposal_serialized":  "0acd010abc0108641a200000000000000000000000000000000000000000000000000000000000000000222000000000000000000000000000000000000000000000000000000000000000002a20000000000000000000000000000000000000000000000000000000000000000032200000000000000000000000000000000000000000000000000000000000000000420e626564726f636b2d746573742d314a20000000000000000000000000000000000000000000000000000000000000000012050a0374783112050a037478321a200000000000000000000000000000000000000000000000000000000000000000224000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	"timeout_serialized":   "083210031a200000000000000000000000000000000000000000000000000000000000000000224000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
}

// golden returns the expected value for a golden vector. On first run (when
// the vector is not yet recorded), it records the value and passes the test.
func golden(t *testing.T, name, got string) string {
	t.Helper()
	want, ok := goldenVectors[name]
	if !ok {
		t.Logf("RECORDING golden vector %q = %q (update goldenVectors map)", name, got)
		return got // First run — accept the value.
	}
	return want
}

