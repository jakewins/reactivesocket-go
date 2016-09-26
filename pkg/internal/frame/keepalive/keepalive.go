package keepalive

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/keepalive"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func Encode(target *frame.Frame, respond bool) *frame.Frame {
	keepalive.Encode(&target.Buf, respond)
	return target
}

func New(respond bool) *frame.Frame {
	f := &frame.Frame{}
	return Encode(f, respond)
}
