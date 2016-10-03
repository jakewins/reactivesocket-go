package errorc

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/errorc"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func ErrorCode(f *frame.Frame) uint32 {
	return errorc.ErrorCode(f.Buf)
}
