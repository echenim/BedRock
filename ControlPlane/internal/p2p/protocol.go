package p2p

import (
	"errors"
	"fmt"

	typesv1 "github.com/echenim/Bedrock/controlplane/gen/proto/bedrock/types/v1"
	"github.com/echenim/Bedrock/controlplane/internal/types"
	"google.golang.org/protobuf/proto"
)

// MessageType identifies the type of consensus message on the wire.
type MessageType byte

const (
	MsgProposal MessageType = 0x01
	MsgVote     MessageType = 0x02
	MsgTimeout  MessageType = 0x03
)

// MaxMessageSize is the maximum allowed message size (4 MB).
// Used as the outer bound in DecodeEnvelope before the type byte is read.
const MaxMessageSize = 4 * 1024 * 1024

// Per-message-type size limits (audit S2).
// Votes and timeouts are small fixed-size messages; allowing 4 MB for them
// enables memory amplification (64 channels × 16 buffer × 4 MB ≈ 4 GB).
const (
	MaxProposalSize = 2 * 1024 * 1024 // 2 MB — proposals contain transactions
	MaxVoteSize     = 4 * 1024        // 4 KB — fixed-size: hash + height + round + sig
	MaxTimeoutSize  = 4 * 1024        // 4 KB — fixed-size: height + round + sig + optional QC ref
)

func (mt MessageType) String() string {
	switch mt {
	case MsgProposal:
		return "proposal"
	case MsgVote:
		return "vote"
	case MsgTimeout:
		return "timeout"
	default:
		return fmt.Sprintf("unknown(0x%02x)", byte(mt))
	}
}

// Envelope wraps a typed message for wire encoding.
type Envelope struct {
	Type    MessageType
	Payload []byte
}

var marshalOpts = proto.MarshalOptions{Deterministic: true}

// Encode serializes the envelope as [type_byte | protobuf_payload].
func (e *Envelope) Encode() []byte {
	buf := make([]byte, 1+len(e.Payload))
	buf[0] = byte(e.Type)
	copy(buf[1:], e.Payload)
	return buf
}

// maxSizeForType returns the per-message-type size limit.
func maxSizeForType(mt MessageType) int {
	switch mt {
	case MsgProposal:
		return MaxProposalSize
	case MsgVote:
		return MaxVoteSize
	case MsgTimeout:
		return MaxTimeoutSize
	default:
		return MaxMessageSize
	}
}

// DecodeEnvelope parses a wire-format message into an Envelope.
// Enforces per-message-type size limits (audit S2).
func DecodeEnvelope(data []byte) (*Envelope, error) {
	if len(data) == 0 {
		return nil, errors.New("p2p: empty message")
	}
	if len(data) > MaxMessageSize {
		return nil, fmt.Errorf("p2p: message too large: %d > %d", len(data), MaxMessageSize)
	}

	mt := MessageType(data[0])
	limit := maxSizeForType(mt)
	if len(data) > limit {
		return nil, fmt.Errorf("p2p: %s message too large: %d > %d", mt, len(data), limit)
	}

	return &Envelope{
		Type:    mt,
		Payload: data[1:],
	}, nil
}

// EncodeProposal serializes a Proposal into wire format.
func EncodeProposal(p *types.Proposal) ([]byte, error) {
	pb := p.ToProto()
	payload, err := marshalOpts.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("p2p: marshal proposal: %w", err)
	}
	env := &Envelope{Type: MsgProposal, Payload: payload}
	return env.Encode(), nil
}

// DecodeProposal deserializes a Proposal from protobuf payload bytes.
func DecodeProposal(payload []byte) (*types.Proposal, error) {
	pb := &typesv1.Proposal{}
	if err := proto.Unmarshal(payload, pb); err != nil {
		return nil, fmt.Errorf("p2p: unmarshal proposal: %w", err)
	}
	return types.ProposalFromProto(pb)
}

// EncodeVote serializes a Vote into wire format.
func EncodeVote(v *types.Vote) ([]byte, error) {
	pb := v.ToProto()
	payload, err := marshalOpts.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("p2p: marshal vote: %w", err)
	}
	env := &Envelope{Type: MsgVote, Payload: payload}
	return env.Encode(), nil
}

// DecodeVote deserializes a Vote from protobuf payload bytes.
func DecodeVote(payload []byte) (*types.Vote, error) {
	pb := &typesv1.Vote{}
	if err := proto.Unmarshal(payload, pb); err != nil {
		return nil, fmt.Errorf("p2p: unmarshal vote: %w", err)
	}
	return types.VoteFromProto(pb)
}

// EncodeTimeout serializes a TimeoutMessage into wire format.
func EncodeTimeout(tm *types.TimeoutMessage) ([]byte, error) {
	pb := tm.ToProto()
	payload, err := marshalOpts.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("p2p: marshal timeout: %w", err)
	}
	env := &Envelope{Type: MsgTimeout, Payload: payload}
	return env.Encode(), nil
}

// DecodeTimeout deserializes a TimeoutMessage from protobuf payload bytes.
func DecodeTimeout(payload []byte) (*types.TimeoutMessage, error) {
	pb := &typesv1.TimeoutMessage{}
	if err := proto.Unmarshal(payload, pb); err != nil {
		return nil, fmt.Errorf("p2p: unmarshal timeout: %w", err)
	}
	return types.TimeoutMessageFromProto(pb)
}

// DecodeMessage decodes a wire-format message into its type and domain object.
// Returns (MessageType, *types.Proposal|*types.Vote|*types.TimeoutMessage, error).
func DecodeMessage(data []byte) (MessageType, interface{}, error) {
	env, err := DecodeEnvelope(data)
	if err != nil {
		return 0, nil, err
	}

	switch env.Type {
	case MsgProposal:
		p, err := DecodeProposal(env.Payload)
		return MsgProposal, p, err
	case MsgVote:
		v, err := DecodeVote(env.Payload)
		return MsgVote, v, err
	case MsgTimeout:
		tm, err := DecodeTimeout(env.Payload)
		return MsgTimeout, tm, err
	default:
		return env.Type, nil, fmt.Errorf("p2p: unknown message type: 0x%02x", byte(env.Type))
	}
}
