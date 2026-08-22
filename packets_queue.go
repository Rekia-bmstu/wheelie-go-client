package wheeliegoclient

import (
	"context"
	"fmt"
)

type packetsQueue struct {
	lowPriorCh    chan packet
	mediumPriorCh chan packet
	highPriorCh   chan packet
	resultCh      chan packet
	queueSize     int
}

func newPacketsQueue(queueSize int) *packetsQueue {
	return &packetsQueue{
		lowPriorCh:    make(chan packet, queueSize),
		mediumPriorCh: make(chan packet, queueSize),
		highPriorCh:   make(chan packet, queueSize),
		resultCh:      make(chan packet), // size = workers count?
		queueSize:     queueSize,
	}
}

func (q *packetsQueue) startQueueReading(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				pkt := nonblockRead(q.highPriorCh, q.mediumPriorCh, q.lowPriorCh)
				if pkt == nil {
					awaited, err := q.waitNext(ctx)
					if err != nil {
						return
					}
					pkt = awaited
				}
				if err := enqueuePacket(ctx, q.resultCh, pkt); err != nil {
					fmt.Println("Error enqueuing packet:", err)
				}
			}
		}
	}()
}

func (q *packetsQueue) waitNext(ctx context.Context) (packet, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case pkt, ok := <-q.highPriorCh:
		if !ok {
			return nil, ctx.Err()
		}
		return pkt, nil
	case pkt, ok := <-q.mediumPriorCh:
		if !ok {
			return nil, ctx.Err()
		}
		return pkt, nil
	case pkt, ok := <-q.lowPriorCh:
		if !ok {
			return nil, ctx.Err()
		}
		return pkt, nil
	}
}

func (q *packetsQueue) push(ctx context.Context, packet packet) error {
	switch packet.Priority() {
	case LowPriority:
		if err := enqueuePacket(ctx, q.lowPriorCh, packet); err != nil {
			return err
		}
	case MediumPriority:
		if err := enqueuePacket(ctx, q.mediumPriorCh, packet); err != nil {
			return err
		}
	case HighPriority:
		if err := enqueuePacket(ctx, q.highPriorCh, packet); err != nil {
			return err
		}
	default:
		panic("unknown packet priority" + string(packet.Priority()))
	}
	return nil
}

func enqueuePacket(ctx context.Context, queue chan packet, packet packet) error {
	select {
	case queue <- packet:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func nonblockRead(channels ...chan packet) packet {
	var result packet = nil
	for _, ch := range channels {
		if result == nil {
			break
		}

		select {
		case result = <-ch:
		default:
		}
	}
	return result
}
