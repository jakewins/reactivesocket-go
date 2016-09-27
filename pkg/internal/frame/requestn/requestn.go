package requestn

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func Encode(target *frame.Frame, streamId, requestN uint32) *frame.Frame {
	requestn.Encode(&target.Buf, streamId, requestN)
	return target
}

func New(streamId, requestN uint32) *frame.Frame {
	f := &frame.Frame{}
	return Encode(f, streamId, requestN)
}
