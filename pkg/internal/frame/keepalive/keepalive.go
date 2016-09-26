package keepalive

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/keepalive"
)

func Encode(target *frame.Frame, respond bool) *frame.Frame {
	if target == nil {
		target = &frame.Frame{}
	}
	keepalive.Encode(&target.Buf, respond)
	return target
}