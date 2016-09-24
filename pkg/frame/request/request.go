package request

import (
	"github.com/jakewins/reactivesocket-go/pkg/frame"
	"github.com/jakewins/reactivesocket-go/pkg/codec/request"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
	"fmt"
)

func Encode(target *frame.Frame, streamId uint32, flags, frameType uint16, metadata, data []byte) {
	request.Encode(&target.Buf, streamId, flags, frameType, metadata, data)
}

func EncodeWithInitialN(target *frame.Frame, streamId, initialN uint32, flags, frameType uint16, metadata, data []byte) {
	request.EncodeWithInitialN(&target.Buf, streamId, initialN, flags, frameType, metadata, data)
}

func InitialRequestN(f *frame.Frame) uint32 {
	switch(f.Type()) {
	case header.FTFireAndForget: return 0;
	case header.FTRequestResponse: return 1;
	case header.FTRequestChannel:
		return request.InitialRequestN(f.Buf)
	case header.FTRequestStream:
		return request.InitialRequestN(f.Buf)
	}
	panic(fmt.Sprintf("Expected a request frame, got %d", f.Type()))
}