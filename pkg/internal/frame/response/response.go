package response

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/response"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func Encode(target *frame.Frame, streamId uint32, flags uint16, metadata, data []byte) *frame.Frame {
	response.Encode(&target.Buf, streamId, flags, metadata, data)
	return target
}

func New(streamId uint32, flags uint16, metadata, data []byte) *frame.Frame {
	f := &frame.Frame{}
	return Encode(f, streamId, flags, metadata, data)
}
