package requestn

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func RequestN(f *frame.Frame) uint32 {
	return requestn.RequestN(f.Buf)
}
