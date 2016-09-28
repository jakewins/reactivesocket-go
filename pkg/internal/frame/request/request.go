package request

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func InitialRequestN(f *frame.Frame) uint32 {
	return request.InitialRequestN(f.Buf)
}
