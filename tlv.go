package wheeliegoclient

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type Tlv struct {
	tcpConn        net.Conn
	baseAddr       string
	receiveBufSize int
	isBigEndian    bool
	maxPacketSize  int32
}

type TlvPacket struct {
	Tag  int32
	Data []byte
}

var (
	ErrNotDialed     = errors.New("not dialed")
	ErrPacketMaxSize = errors.New("packet max size")
)

func NewTlv(host string, port uint16) *Tlv {
	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
	return &Tlv{
		tcpConn:        nil,
		baseAddr:       addr,
		receiveBufSize: 4096,
		isBigEndian:    true,
		maxPacketSize:  16 * 1024,
	}
}

func (t *Tlv) Dial(ctx context.Context) error {
	var d net.Dialer
	tcpConn, err := d.DialContext(ctx, "tcp", t.baseAddr)
	if err != nil {
		return fmt.Errorf("dial tcp: %w", err)
	}

	t.tcpConn = tcpConn
	return nil
}

func (t *Tlv) Close() error {
	if t.tcpConn == nil {
		return ErrNotDialed
	}

	if err := t.tcpConn.Close(); err != nil {
		return fmt.Errorf("close tcp: %w", err)
	}
	return nil
}

func (t *Tlv) Send(ctx context.Context, tag int32, data []byte) error {
	if t.tcpConn == nil {
		return ErrNotDialed
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := t.tcpConn.SetWriteDeadline(deadline); err != nil {
			return err
		}
		defer t.tcpConn.SetWriteDeadline(time.Time{})
	}

	if err := t.validatePacketLength(int32(len(data))); err != nil {
		return err
	}

	packet := &TlvPacket{
		Tag:  tag,
		Data: data,
	}
	pktBytes, err := packet.MarshalBinary(t.endian())
	if err != nil {
		return fmt.Errorf("marshal packet: %w", err)
	}
	_, err = t.tcpConn.Write(pktBytes)
	if err != nil {
		return fmt.Errorf("write packet: %w", err)
	}
	return nil
}

func (t *Tlv) StartReceive(ctx context.Context) (<-chan *TlvPacket, <-chan error, error) {
	if t.tcpConn == nil {
		return nil, nil, ErrNotDialed
	}

	result := make(chan *TlvPacket, t.receiveBufSize)
	errc := make(chan error, 1)

	go func() {
		defer close(result)
		defer close(errc)

		if deadline, ok := ctx.Deadline(); ok {
			if err := t.tcpConn.SetReadDeadline(deadline); err != nil {
				errc <- fmt.Errorf("set read deadline: %w", err)
				return
			}
			defer t.tcpConn.SetReadDeadline(time.Time{})
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				tagBuf := make([]byte, 4)
				_, err := io.ReadFull(t.tcpConn, tagBuf)
				if err != nil {
					errc <- fmt.Errorf("read tag: %w", err)
					return
				}

				lenBuf := make([]byte, 4)
				_, err = io.ReadFull(t.tcpConn, lenBuf)
				if err != nil {
					errc <- fmt.Errorf("read len: %w", err)
					return
				}
				length := int32(t.endian().Uint32(lenBuf))
				if err := t.validatePacketLength(length); err != nil {
					errc <- err
					return
				}

				data := make([]byte, length)
				_, err = io.ReadFull(t.tcpConn, data)
				if err != nil {
					errc <- fmt.Errorf("read data: %w", err)
					return
				}

				packet := &TlvPacket{
					Tag:  int32(t.endian().Uint32(tagBuf)),
					Data: data,
				}
				select {
				case <-ctx.Done():
					return
				case result <- packet:
				}
			}
		}
	}()

	return result, errc, nil
}

func (t *Tlv) endian() binary.ByteOrder {
	if t.isBigEndian {
		return binary.BigEndian
	}

	return binary.LittleEndian
}

func (t *Tlv) validatePacketLength(length int32) error {
	if length > t.maxPacketSize || length < 0 {
		return ErrPacketMaxSize
	}

	return nil
}

func (p *TlvPacket) MarshalBinary(endian binary.ByteOrder) ([]byte, error) {
	var buf bytes.Buffer

	tag := p.Tag
	err := binary.Write(&buf, endian, tag)
	if err != nil {
		return nil, fmt.Errorf("write tag: %w", err)
	}

	length := len(p.Data)
	err = binary.Write(&buf, endian, int32(length))
	if err != nil {
		return nil, fmt.Errorf("write length: %w", err)
	}

	data := p.Data
	err = binary.Write(&buf, endian, data)
	if err != nil {
		return nil, fmt.Errorf("write data: %w", err)
	}
	return buf.Bytes(), nil
}
