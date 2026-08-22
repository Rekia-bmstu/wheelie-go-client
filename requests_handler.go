package wheeliegoclient

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type request struct {
	timestamp         time.Time
	responseTimestamp *time.Time
	responseCh        chan packet
	ctx               context.Context
}

type requestsHandler struct {
	requests map[uuid.UUID]request
}

func newRequestsHandler() *requestsHandler {
	return &requestsHandler{
		requests: make(map[uuid.UUID]request),
	}
}

func (rh *requestsHandler) add(ctx context.Context, requestID uuid.UUID) {
	req := request{
		timestamp:         time.Now().UTC(),
		responseCh:        make(chan packet, 1),
		ctx:               ctx,
		responseTimestamp: nil,
	}
	rh.requests[requestID] = req
}

func (rh *requestsHandler) wait(ctx context.Context, requestID uuid.UUID) (packet, error) {
	req, ok := rh.requests[requestID]
	if !ok {
		return nil, &RequestNotFoundError{
			requestID: requestID,
		}
	}
	select {
	case <-req.ctx.Done():
		return nil, req.ctx.Err()
	case <-ctx.Done():
		return nil, req.ctx.Err()
	case resp := <-req.responseCh:
		return resp, nil
	}
}

func (rh *requestsHandler) dropAll() {
	for _, req := range rh.requests {
		close(req.responseCh)
	}
}

func waitTyped[T packet](ctx context.Context, rh *requestsHandler, requestID uuid.UUID) (T, error) {
	var dummy T
	respPkt, err := rh.wait(ctx, requestID)
	if err != nil {
		return dummy, err
	}
	typed, ok := respPkt.(T)
	if !ok {
		return dummy, err
	}
	return typed, nil
}

func (rh *requestsHandler) ack(ctx context.Context, requestID uuid.UUID, resp packet) error {
	req, ok := rh.requests[requestID]
	if !ok {
		return &RequestNotFoundError{
			requestID: requestID,
		}
	}
	req.responseTimestamp = new(time.Now().UTC())
	select {
	case <-ctx.Done():
		return ctx.Err()
	case req.responseCh <- resp:
		return nil
	}
}
