package wheeliegoclient

import (
	"fmt"
	"wheelie-client/internal/pb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func toConnectResponsePacket(tlvPkt *TlvPacket) (*connectResponsePacket, error) {
	if tlvPkt.Tag != int32(ConnectResponse) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}

	pbPkt := pb.ConnectResponsePacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}

	serverID, err := uuid.Parse(pbPkt.ServerId)
	if err != nil {
		return nil, err
	}

	return &connectResponsePacket{
		serverID: serverID,
		code:     ConnectStatusCode(pbPkt.StatusCode),
		desc:     pbPkt.Description,
	}, nil
}

func toSetupResponsePacket(tlvPkt *TlvPacket) (*setupResponsePacket, error) {
	if tlvPkt.Tag != int32(SetupResponse) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}

	pbPkt := pb.SetupResponsePacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}

	pbPayload := pb.SetupResponsePacketPayload{}
	err = proto.Unmarshal(pbPkt.Payload, &pbPayload)
	if err != nil {
		return nil, err
	}
	return &setupResponsePacket{
		needAuth:  pbPayload.NeedAuth,
		canResume: pbPkt.CanResume,
		payload:   pbPkt.Payload,
	}, nil
}

func toAuthResponsePacket(tlvPkt *TlvPacket) (*authResponsePacket, error) {
	if tlvPkt.Tag != int32(AuthResponse) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}
	pbPkt := pb.AuthResponsePacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}
	return &authResponsePacket{
		code:    AuthStatusCode(pbPkt.StatusCode),
		desc:    pbPkt.Description,
		payload: pbPkt.Payload,
	}, nil
}

func toPingResponsePacket(tlvPkt *TlvPacket) (*pingResponsePacket, error) {
	if tlvPkt.Tag != int32(PingResponse) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}
	pbPkt := pb.PingResponsePacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}
	return &pingResponsePacket{
		requestID: uuid.MustParse(pbPkt.RequestId),
	}, nil
}

func toServerPushAckPacket(tlvPkt *TlvPacket) (*serverPushAckPacket, error) {
	if tlvPkt.Tag != int32(ServerPushAck) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}

	pbPkt := pb.ServerPushAckPacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}
	return &serverPushAckPacket{requestID: uuid.MustParse(pbPkt.RequestId)}, nil
}

func toServerPushPacket(tlvPkt *TlvPacket) (*serverPushPacket, error) {
	if tlvPkt.Tag != int32(ServerPush) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}
	pbPkt := pb.ServerPushPacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}
	return &serverPushPacket{
		requestID: uuid.MustParse(pbPkt.Id),
		pktType:   pbPkt.Type,
		payload:   pbPkt.Payload,
	}, nil
}

func toGetResponsePacket(tlvPkt *TlvPacket) (*getResponsePacket, error) {
	if tlvPkt.Tag != int32(GetResponse) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}

	pbPkt := pb.GetResponsePacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}
	return &getResponsePacket{
		requestID: uuid.MustParse(pbPkt.RequestId),
		pktType:   pbPkt.Type,
		payload:   pbPkt.Payload,
		code:      GetStatusCode(pbPkt.StatusCode),
		desc:      pbPkt.Description,
	}, nil
}

func toFireAndForgetPacket(tlvPkt *TlvPacket) (*fireAndForgetPacket, error) {
	if tlvPkt.Tag != int32(FireAndForget) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}
	pbPkt := pb.FireAndForgetPacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}
	return newFireAndForgetPacket(uuid.MustParse(pbPkt.Id), pbPkt.Type, pbPkt.Payload), nil
}

func toDisconnectPacket(tlvPkt *TlvPacket) (*disconnectPacket, error) {
	if tlvPkt.Tag != int32(Disconnect) {
		return nil, fmt.Errorf("%w:%d", ErrWrongPacketTag, tlvPkt.Tag)
	}
	pbPkt := pb.DisconnectPacket{}
	err := proto.Unmarshal(tlvPkt.Data, &pbPkt)
	if err != nil {
		return nil, err
	}
	return newDisconnectPacket(DisconnectStatusCode(pbPkt.StatusCode), pbPkt.Description), nil
}
