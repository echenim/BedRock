package types

import (
	"testing"

	typesv1 "github.com/echenim/Bedrock/controlplane/gen/proto/bedrock/types/v1"
	"google.golang.org/protobuf/proto"
)

// FuzzBlockFromProto fuzzes the Block protobuf decoder.
func FuzzBlockFromProto(f *testing.F) {
	// Seed with a minimal valid proto.
	valid := &typesv1.Block{Header: &typesv1.BlockHeader{Height: 1}}
	if b, err := proto.Marshal(valid); err == nil {
		f.Add(b)
	}
	f.Add([]byte{})
	f.Add([]byte{0x0a, 0x00})

	f.Fuzz(func(t *testing.T, data []byte) {
		var pb typesv1.Block
		if err := proto.Unmarshal(data, &pb); err != nil {
			return
		}
		_, _ = BlockFromProto(&pb)
	})
}

// FuzzVoteFromProto fuzzes the Vote protobuf decoder.
func FuzzVoteFromProto(f *testing.F) {
	valid := &typesv1.Vote{Height: 1, Round: 0}
	if b, err := proto.Marshal(valid); err == nil {
		f.Add(b)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var pb typesv1.Vote
		if err := proto.Unmarshal(data, &pb); err != nil {
			return
		}
		_, _ = VoteFromProto(&pb)
	})
}

// FuzzProposalFromProto fuzzes the Proposal protobuf decoder.
func FuzzProposalFromProto(f *testing.F) {
	valid := &typesv1.Proposal{Round: 1}
	if b, err := proto.Marshal(valid); err == nil {
		f.Add(b)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var pb typesv1.Proposal
		if err := proto.Unmarshal(data, &pb); err != nil {
			return
		}
		_, _ = ProposalFromProto(&pb)
	})
}

// FuzzQCFromProto fuzzes the QuorumCertificate protobuf decoder.
func FuzzQCFromProto(f *testing.F) {
	valid := &typesv1.QuorumCertificate{Round: 1}
	if b, err := proto.Marshal(valid); err == nil {
		f.Add(b)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var pb typesv1.QuorumCertificate
		if err := proto.Unmarshal(data, &pb); err != nil {
			return
		}
		_, _ = QCFromProto(&pb)
	})
}
