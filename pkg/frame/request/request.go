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

func InitialRequestN(f *frame.Frame) int {
	switch(f.Type()) {
	case header.FTFireAndForget: return 0;
	case header.FTRequestResponse: return 1;
	}
	panic(fmt.Sprintf("Expected a request frame, got %d", f.Type()))
}