package wheeliegoclient

import (
	"fmt"
	"wheelie-client/internal/pb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type packet interface {
	Tag() PacketTag
	Priority() Priority
	ToTLV() (*TlvPacket, error)
}

type connectPacket struct {
	tag      PacketTag
	clientID uuid.UUID
}

type connectResponsePacket struct {
	serverID uuid.UUID
	code     ConnectStatusCode
	desc     string
}

func (c *connectResponsePacket) Tag() PacketTag {
	//TODO implement me
	panic("implement me")
}

func (c *connectResponsePacket) Priority() Priority {
	//TODO implement me
	panic("implement me")
}

func (c *connectResponsePacket) ToTLV() (*TlvPacket, error) {
	//TODO implement me
	panic("implement me")
}

func newConnectPacket(clientID uuid.UUID) *connectPacket {
	return &connectPacket{
		tag:      Connect,
		clientID: clientID,
	}
}

func (p *connectPacket) Priority() Priority {
	return MediumPriority
}

func (p *connectPacket) Tag() PacketTag {
	return p.tag
}

func (p *connectPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.ConnectPacket{ClientId: p.clientID.String()}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}

	return &TlvPacket{
		Tag:  int32(p.tag),
		Data: bytes,
	}, nil
}

type setupPacket struct {
	protocolVersion string
	payload         []byte
}

func newSetupPacket(protocolVersion string, payload []byte) *setupPacket {
	return &setupPacket{
		protocolVersion: protocolVersion,
		payload:         payload,
	}
}

func (s *setupPacket) Tag() PacketTag {
	return Setup
}

func (s *setupPacket) Priority() Priority {
	return MediumPriority
}

func (s *setupPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.SetupPacket{
		ProtocolVersion: s.protocolVersion,
		Payload:         s.payload,
	}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{
		Tag:  int32(Setup),
		Data: bytes,
	}, nil
}

type setupResponsePacket struct {
	needAuth  bool
	canResume bool
	payload   []byte
}

func (p *setupResponsePacket) Tag() PacketTag {
	//TODO implement me
	panic("implement me")
}

func (p *setupResponsePacket) Priority() Priority {
	//TODO implement me
	panic("implement me")
}

func (p *setupResponsePacket) ToTLV() (*TlvPacket, error) {
	//TODO implement me
	panic("implement me")
}

type authPacket struct {
	data []byte
}

func newAuthPacket(data []byte) (*authPacket, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}
	return &authPacket{
		data: data,
	}, nil
}

func (a *authPacket) Tag() PacketTag {
	return Auth
}

func (a *authPacket) Priority() Priority {
	return MediumPriority
}

func (a *authPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.AuthPacket{
		Payload: a.data,
	}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{
		Tag:  int32(a.Tag()),
		Data: bytes,
	}, nil
}

type authResponsePacket struct {
	code    AuthStatusCode
	desc    string
	payload []byte
}

func (a *authResponsePacket) Tag() PacketTag {
	//TODO implement me
	panic("implement me")
}

func (a *authResponsePacket) Priority() Priority {
	//TODO implement me
	panic("implement me")
}

func (a *authResponsePacket) ToTLV() (*TlvPacket, error) {
	//TODO implement me
	panic("implement me")
}

type pingPacket struct {
	requestID uuid.UUID
}

func newPingPacket(requestID uuid.UUID) *pingPacket {
	return &pingPacket{
		requestID: requestID,
	}
}

func (p *pingPacket) Tag() PacketTag {
	return Ping
}

func (p *pingPacket) Priority() Priority {
	return MediumPriority
}

func (p *pingPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.PingPacket{
		Id: p.requestID.String(),
	}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{Tag: int32(p.Tag()), Data: bytes}, nil
}

type pingResponsePacket struct {
	requestID uuid.UUID
}

func newPingResponsePacket(requestID uuid.UUID) *pingResponsePacket {
	return &pingResponsePacket{
		requestID: requestID,
	}
}

func (p *pingResponsePacket) Tag() PacketTag {
	//TODO implement me
	panic("implement me")
}

func (p *pingResponsePacket) Priority() Priority {
	//TODO implement me
	panic("implement me")
}

func (p *pingResponsePacket) ToTLV() (*TlvPacket, error) {
	//TODO implement me
	panic("implement me")
}

type serverPushPacket struct {
	requestID uuid.UUID
	pktType   int32
	payload   []byte
}

func newServerPushPacket(requestID uuid.UUID, pktType int32, payload []byte) *serverPushPacket {
	return &serverPushPacket{
		requestID: requestID,
		pktType:   pktType,
		payload:   payload,
	}
}

func (s *serverPushPacket) Tag() PacketTag {
	return ServerPush
}

func (s *serverPushPacket) Priority() Priority {
	return MediumPriority
}

func (s *serverPushPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.ServerPushPacket{
		Id:      s.requestID.String(),
		Type:    s.pktType,
		Payload: s.payload,
	}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{Tag: int32(s.Tag()), Data: bytes}, nil
}

type serverPushAckPacket struct {
	requestID uuid.UUID
}

func newServerPushAckPacket(requestID uuid.UUID) *serverPushAckPacket {
	return &serverPushAckPacket{
		requestID: requestID,
	}
}

func (s *serverPushAckPacket) Tag() PacketTag {
	return ServerPushAck
}

func (s *serverPushAckPacket) Priority() Priority {
	return MediumPriority
}

func (s *serverPushAckPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.ServerPushAckPacket{}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{Tag: int32(s.Tag()), Data: bytes}, nil
}

type getPacket struct {
	requestID uuid.UUID
	pktType   int32
	payload   []byte
}

func newGetPacket(requestID uuid.UUID, pktType int32, payload []byte) *getPacket {
	return &getPacket{
		requestID: requestID,
		pktType:   pktType,
		payload:   payload,
	}
}

func (p *getPacket) Tag() PacketTag {
	return Get
}

func (p *getPacket) Priority() Priority {
	return MediumPriority
}

func (p *getPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.GetPacket{
		Id:      p.requestID.String(),
		Type:    p.pktType,
		Payload: p.payload,
	}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{Tag: int32(p.Tag()), Data: bytes}, nil
}

type getResponsePacket struct {
	requestID uuid.UUID
	pktType   int32
	payload   []byte
	code      GetStatusCode
	desc      string
}

func (p *getResponsePacket) Tag() PacketTag {
	//TODO implement me
	panic("implement me")
}

func (p *getResponsePacket) Priority() Priority {
	//TODO implement me
	panic("implement me")
}

func (p *getResponsePacket) ToTLV() (*TlvPacket, error) {
	//TODO implement me
	panic("implement me")
}

type closePacket struct {
	code CloseStatusCode
	desc string
}

func newClosePacket(code CloseStatusCode, desc string) *closePacket {
	return &closePacket{
		code: code,
		desc: desc,
	}
}

func (p *closePacket) Tag() PacketTag {
	return Close
}

func (p *closePacket) Priority() Priority {
	return HighPriority
}

func (p *closePacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.ClosePacket{}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{Tag: int32(p.Tag()), Data: bytes}, nil
}

type fireAndForgetPacket struct {
	id      uuid.UUID
	pktType int32
	payload []byte
}

func newFireAndForgetPacket(id uuid.UUID, pktType int32, payload []byte) *fireAndForgetPacket {
	return &fireAndForgetPacket{
		id:      id,
		pktType: pktType,
		payload: payload,
	}
}

func (p *fireAndForgetPacket) Tag() PacketTag {
	return FireAndForget
}

func (p *fireAndForgetPacket) Priority() Priority {
	return MediumPriority
}

func (p *fireAndForgetPacket) ToTLV() (*TlvPacket, error) {
	pbPkt := pb.FireAndForgetPacket{
		Id:      p.id.String(),
		Type:    p.pktType,
		Payload: p.payload,
	}
	bytes, err := proto.Marshal(&pbPkt)
	if err != nil {
		return nil, err
	}
	return &TlvPacket{Tag: int32(p.Tag()), Data: bytes}, nil
}

type disconnectPacket struct {
	code DisconnectStatusCode
	desc string
}

func newDisconnectPacket(code DisconnectStatusCode, desc string) *disconnectPacket {
	return &disconnectPacket{
		code: code,
		desc: desc,
	}
}
