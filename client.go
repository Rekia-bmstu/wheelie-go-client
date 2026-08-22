package wheeliegoclient

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)
import "github.com/google/uuid"

type Client struct {
	tlv             *Tlv
	logger          *slog.Logger
	connectionID    uuid.UUID
	pingInterval    time.Duration
	utilID          uuid.UUID
	reqHandler      *requestsHandler
	pktQueue        *packetsQueue
	fireAndForgetCh chan *WheelieData
	serverPushCh    chan *WheelieData
}

func NewClient() *Client {
	logger := slog.Default()

	return &Client{
		logger:          logger,
		connectionID:    uuid.New(),
		pingInterval:    10,
		utilID:          uuid.New(),
		reqHandler:      newRequestsHandler(),
		pktQueue:        newPacketsQueue(100),
		fireAndForgetCh: make(chan *WheelieData, 100),
		serverPushCh:    make(chan *WheelieData, 100),
	}
}

func (c *Client) Connect(ctx context.Context, host string, port uint16, authData []byte) error {
	tlv := NewTlv(host, port)
	c.tlv = tlv
	err := c.tlv.Dial(ctx)
	if err != nil {
		return fmt.Errorf("%w:%w", ErrConnectAsync, err)
	}

	if err = c.startListening(ctx); err != nil {
		return fmt.Errorf("%w:%w", ErrConnectAsync, err)
	}
	if err = c.startSending(ctx); err != nil {
		return fmt.Errorf("%w:%w", ErrConnectAsync, err)
	}

	respConnPkt, err := c.sendConnect(ctx)
	if err != nil {
		return err
	}

	if respConnPkt.code != SuccessConnect {
		return ErrCouldntConnect
	}
	setupRespPkt, err := c.sendSetup(ctx)
	if err != nil {
		return err
	}
	if setupRespPkt.needAuth {
		authRespPkt, err := c.sendAuth(ctx, authData)
		if err != nil {
			return err
		}
		if authRespPkt.code != SuccessAuth {
			return ErrCouldntConnect
		}
	}

	c.startPingRoutine(ctx)
	return nil
}

func (c *Client) Get(ctx context.Context, data *WheelieData) (*WheelieData, error) {
	pkt := newGetPacket(
		uuid.New(),
		data.PktType,
		data.Data,
	)
	resp, err := sendRequestPacket[*getResponsePacket](ctx, c, pkt.requestID, pkt)
	if err != nil {
		return nil, err
	}
	if resp.code != SuccessGet {
		return nil, fmt.Errorf("%w:%d:%s", ErrCouldntConnect, resp.code, resp.desc)
	}
	return &WheelieData{
		resp.pktType,
		resp.payload,
	}, nil
}

func (c *Client) FireAndForget(ctx context.Context, data *WheelieData) error {
	pkt := newFireAndForgetPacket(
		uuid.New(),
		data.PktType,
		data.Data,
	)
	if err := c.sendPacket(ctx, pkt); err != nil {
		return err
	}
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	pkt := newClosePacket(CloseClientAppShutdown, "")
	err := c.sendPacketForce(ctx, pkt)
	if err != nil {
		return err
	}

	if err := c.dispose(); err != nil {
		return err
	}
	return nil
}

func (c *Client) dispose() error {
	if err := c.tlv.Close(); err != nil {
		return err
	}

	c.reqHandler.dropAll()
	return nil
}

func (c *Client) GetFireAndForgetChannel() <-chan *WheelieData {
	return c.fireAndForgetCh
}

func (c *Client) GetServerPushChannel() <-chan *WheelieData {
	return c.serverPushCh
}

func (c *Client) sendConnect(ctx context.Context) (*connectResponsePacket, error) {
	pkt := newConnectPacket(c.connectionID)
	connRespPkt, err := sendRequestPacket[*connectResponsePacket](ctx, c, c.utilID, pkt)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrConnectAsync, err)
	}

	return connRespPkt, nil
}

func (c *Client) sendSetup(ctx context.Context) (*setupResponsePacket, error) {
	pkt := newSetupPacket("1.0", nil)
	setupRespPkt, err := sendRequestPacket[*setupResponsePacket](ctx, c, c.utilID, pkt)
	if err != nil {
		return nil, err
	}

	return setupRespPkt, nil
}

func (c *Client) sendAuth(ctx context.Context, authData []byte) (*authResponsePacket, error) {
	authPkt, err := newAuthPacket(authData)
	if err != nil {
		return nil, err
	}
	authRespPkt, err := sendRequestPacket[*authResponsePacket](ctx, c, c.utilID, authPkt)
	if err != nil {
		return nil, err
	}
	return authRespPkt, nil
}

func (c *Client) sendPacketForce(ctx context.Context, pkt packet) error {
	tlvPkt, err := pkt.ToTLV()
	if err != nil {
		return err
	}
	err = c.tlv.Send(ctx, int32(pkt.Tag()), tlvPkt.Data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) sendPacket(ctx context.Context, pkt packet) error {
	err := c.pktQueue.push(ctx, pkt)
	if err != nil {
		return err
	}

	return nil
}

func sendRequestPacket[TResponse packet](ctx context.Context, client *Client, requestID uuid.UUID, pkt packet) (TResponse, error) {
	var dummy TResponse
	client.reqHandler.add(ctx, requestID)

	if err := client.sendPacket(ctx, pkt); err != nil {
		return dummy, err
	}

	respPkt, waitErr := waitTyped[TResponse](ctx, client.reqHandler, requestID)
	if waitErr != nil {
		return dummy, waitErr
	}
	return respPkt, nil
}

func (c *Client) startListening(ctx context.Context) error {
	pkts, done, err := c.tlv.StartReceive(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case pkt, ok := <-pkts:
				if !ok {
					return
				}
				if err := c.handleReceivedPacket(ctx, pkt); err != nil {
					return
				}
			}
		}

	}()
	return nil
}

func (c *Client) handleReceivedPacket(ctx context.Context, pkt *TlvPacket) error {
	switch PacketTag(pkt.Tag) {
	case ConnectResponse:
		connRespPkt, err := toConnectResponsePacket(pkt)
		if err != nil {
			return err
		}
		c.onConnectResponsePacket(ctx, connRespPkt)
	case SetupResponse:
		setupRespPkt, err := toSetupResponsePacket(pkt)
		if err != nil {
			return err
		}
		c.onSetupResponsePacket(ctx, setupRespPkt)
	case AuthResponse:
		authRespPkt, err := toAuthResponsePacket(pkt)
		if err != nil {
			return err
		}
		c.onAuthStatusPacket(ctx, authRespPkt)
	case PingResponse:
		pong, err := toPingResponsePacket(pkt)
		if err != nil {
			return err
		}
		c.onPingAckPacket(ctx, pong)
	case GetResponse:
		getRespPkt, err := toGetResponsePacket(pkt)
		if err != nil {
			return err
		}
		c.onGetResponsePacket(ctx, getRespPkt)
	case FireAndForget:
		fireAndForgetPkt, err := toFireAndForgetPacket(pkt)
		if err != nil {
			return err
		}
		c.onFireAndForgetPacket(ctx, fireAndForgetPkt)
	case ServerPush:
		serverPushPkt, err := toServerPushPacket(pkt)
		if err != nil {
			return err
		}
		if err := c.onServerPushPacket(ctx, serverPushPkt); err != nil {
			return err
		}
	case Disconnect:
		disconnPkt, err := toDisconnectPacket(pkt)
		if err != nil {
			return err
		}
		if err := c.onDisconnectPacket(ctx, disconnPkt); err != nil {
			return err
		}
	default:
		fmt.Println("unknown packet tag: ", pkt.Tag)
	}
	return nil
}

func (c *Client) startSending(ctx context.Context) error {
	c.pktQueue.startQueueReading(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case pkt, ok := <-c.pktQueue.resultCh:
				if !ok {
					return
				}
				if err := c.sendPacketForce(ctx, pkt); err != nil {
					return
				}
			}
		}
	}()

	return nil
}

func (c *Client) startPingRoutine(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Second * c.pingInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*c.pingInterval)
				ping := newPingPacket(c.utilID)
				_, err := sendRequestPacket[*pingResponsePacket](timeoutCtx, c, c.utilID, ping)
				cancel()
				if err != nil {
					// Disconnect
					return
				}
			}
		}
	}()
}

func (c *Client) onConnectResponsePacket(ctx context.Context, connRespPkt *connectResponsePacket) {
	err := c.reqHandler.ack(ctx, c.utilID, connRespPkt)
	if err != nil {
		return
	}
}

func (c *Client) onSetupResponsePacket(ctx context.Context, setupRespPkt *setupResponsePacket) {
	err := c.reqHandler.ack(ctx, c.utilID, setupRespPkt)
	if err != nil {
		return
	}
}

func (c *Client) onAuthStatusPacket(ctx context.Context, authRespPkt *authResponsePacket) {
	err := c.reqHandler.ack(ctx, c.utilID, authRespPkt)
	if err != nil {
		return
	}
}

func (c *Client) onPingAckPacket(ctx context.Context, pingPkt *pingResponsePacket) {
	err := c.reqHandler.ack(ctx, pingPkt.requestID, pingPkt)
	if err != nil {
		return
	}
}

func (c *Client) onGetResponsePacket(ctx context.Context, getRespPkt *getResponsePacket) {
	err := c.reqHandler.ack(ctx, getRespPkt.requestID, getRespPkt)
	if err != nil {
		return
	}
}

func (c *Client) onFireAndForgetPacket(ctx context.Context, fireAndForgetPacket *fireAndForgetPacket) {
	select {
	case <-ctx.Done():
		return
	case c.fireAndForgetCh <- wheelieDataFromFireAndForgetPacket(fireAndForgetPacket):
	}
}

func (c *Client) onServerPushPacket(ctx context.Context, serverPushPkt *serverPushPacket) error {
	if err := c.sendPacket(ctx, newServerPushAckPacket(serverPushPkt.requestID)); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.serverPushCh <- NewWheelieData(serverPushPkt.pktType, serverPushPkt.payload):
		return nil
	}
}

func (c *Client) onDisconnectPacket(ctx context.Context, disconnPkt *disconnectPacket) error {
	return c.dispose()
}

type WheelieData struct {
	PktType int32
	Data    []byte
}

func NewWheelieData(pktType int32, data []byte) *WheelieData {
	return &WheelieData{
		PktType: pktType,
		Data:    data,
	}
}

func wheelieDataFromFireAndForgetPacket(pkt *fireAndForgetPacket) *WheelieData {
	return NewWheelieData(pkt.pktType, pkt.payload)
}

// <3
//#include <stdio.h>
//int main(void)
//{
//printf("Hello, World! \n" );
//return 0;
//}
