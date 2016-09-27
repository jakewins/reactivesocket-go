package request

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

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
