package response

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
)

func Encode(bufPtr *[]byte, streamId uint32, flags uint16, metadata, data []byte) {
	buf := header.ResizeSlice(bufPtr, header.ComputeLength(len(metadata), len(data)))
	if len(metadata) > 0 {
		flags |= header.FlagHasMetadata
	}

	header.EncodeHeader(buf, flags, header.FTResponse, streamId)
	header.EncodeMetaDataAndData(buf, metadata, data, header.FrameHeaderLength, flags)
}

func PayloadOffset() int {
	return header.FrameHeaderLength
}

func Describe(buf []byte) string {
	return fmt.Sprintf("Response{streamId=%d, metadata=[% x], data=[% x]}",
		header.StreamID(buf), header.Metadata(buf, PayloadOffset), header.Data(buf, PayloadOffset))
}
