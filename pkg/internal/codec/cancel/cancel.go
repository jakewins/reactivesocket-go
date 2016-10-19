package cancel

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
)

func Encode(bufPtr *[]byte, streamId uint32) {
	buf := header.ResizeSlice(bufPtr, header.ComputeLength(0, 0))
	header.EncodeHeader(buf, 0, header.FTCancel, streamId)
}

func PayloadOffset() int {
	return header.FrameHeaderLength
}

func Describe(buf []byte) string {
	return fmt.Sprintf("Cancel{streamId=%d}", header.StreamID(buf))
}
