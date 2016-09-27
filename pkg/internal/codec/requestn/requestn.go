package requestn

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
)

const (
	sizeOfInt           = header.SizeOfInt
	requestNFieldOffset = header.FrameHeaderLength
)

func computeFrameLength() int {
	length := header.ComputeLength(0, 0)
	length += sizeOfInt
	return length
}

func Encode(bufPtr *[]byte, streamId, requestN uint32) {
	buf := header.ResizeSlice(bufPtr, computeFrameLength())

	header.EncodeHeader(buf, 0, header.FTRequestN, streamId)

	header.PutUint32(buf, requestNFieldOffset, requestN)

	offset := requestNFieldOffset
	offset += sizeOfInt
}

func RequestN(b []byte) uint32 {
	return header.Uint32(b, requestNFieldOffset)
}

func Describe(buf []byte) string {
	return fmt.Sprintf("RequestN{streamId=%d, N=%d}", header.StreamID(buf), RequestN(buf))
}

func PayloadOffset() int {
	return requestNFieldOffset
}
