package request

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func Encode(target *frame.Frame, streamId uint32, flags, frameType uint16, metadata, data []byte) *frame.Frame {
	request.Encode(&target.Buf, streamId, flags, frameType, metadata, data)
	return target
}

func EncodeWithInitialN(target *frame.Frame, streamId, initialN uint32, flags, frameType uint16, metadata, data []byte) *frame.Frame {
	request.EncodeWithInitialN(&target.Buf, streamId, initialN, flags, frameType, metadata, data)
	return target
}

func New(streamId uint32, flags, frameType uint16, metadata, data []byte) *frame.Frame {
	f := &frame.Frame{}
	return Encode(f, streamId, flags, frameType, metadata, data)
}

func NewWithInitialN(streamId, initialN uint32, flags, frameType uint16, metadata, data []byte) *frame.Frame {
	f := &frame.Frame{}
	return EncodeWithInitialN(f, streamId, initialN, flags, frameType, metadata, data)
}

func InitialRequestN(f *frame.Frame) uint32 {
	switch f.Type() {
	case header.FTFireAndForget:
		return 0
	case header.FTRequestResponse:
		return 1
	case header.FTRequestChannel:
		return request.InitialRequestN(f.Buf)
	case header.FTRequestStream:
		return request.InitialRequestN(f.Buf)
	case header.FTRequestSubscription:
		return request.InitialRequestN(f.Buf)
	}
	panic(fmt.Sprintf("Expected a request frame, got %d", f.Type()))
}