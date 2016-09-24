package request

import (
"github.com/jakewins/reactivesocket-go/pkg/codec/header"
)

const (
	sizeOfInt = header.SizeOfInt
	initialNFieldOffset = header.FrameHeaderLength
)

func computeFrameLengthWithInitialN(metadata, data []byte) int {
	length := header.ComputeLength(len(metadata), len(data))
	length += sizeOfInt
	return length
}

func Encode(bufPtr *[]byte, streamId uint32, flags, frameType uint16, metadata, data []byte) {
	buf := header.ResizeSlice(bufPtr, header.ComputeLength(len(metadata), len(data)))
	if len(metadata) > 0 {
		flags |= header.FlagHasMetadata
	}

	header.EncodeHeader(buf, flags, frameType, streamId)
	header.EncodeMetaDataAndData(buf, metadata, data, header.FrameHeaderLength, flags)
}

func EncodeWithInitialN(bufPtr *[]byte, streamId, initialN uint32, flags, frameType uint16, metadata, data []byte) {
	buf := header.ResizeSlice(bufPtr, computeFrameLengthWithInitialN(metadata, data))
	if len(metadata) > 0 {
		flags |= header.FlagHasMetadata
	}

	header.EncodeHeader(buf, flags, frameType, 0)
}

func PayloadOffset(b []byte) int {
	return header.FrameHeaderLength
}

func PayloadOffsetWithInitialN(b []byte) int {
	return initialNFieldOffset + sizeOfInt
}