package wheeliegoclient

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrConnectAsync = errors.New("connect async")
var ErrWrongPacketTag = errors.New("wrong packet tag")
var ErrCouldntConnect = errors.New("couldn't connect")

type RequestNotFoundError struct {
	requestID uuid.UUID
}

func (e *RequestNotFoundError) Error() string {
	return fmt.Sprintf("request %s not found", e.requestID)
}

type WrongPacketTagError struct {
	tag string
}

func (e *WrongPacketTagError) Error() string {
	return fmt.Sprintf("wrong packet tag %s", e.tag)
}
