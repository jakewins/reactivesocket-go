package request

import (
"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
)

const (
	RequestFlagRequestChannelC = 1 << 12
	RequestFlagRequestChannelN = 1 << 11
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
	flags |= RequestFlagRequestChannelN

	header.EncodeHeader(buf, flags, frameType, streamId)

	header.PutUint32(buf, initialNFieldOffset, initialN)

	offset := initialNFieldOffset
	offset += sizeOfInt

	header.EncodeMetaDataAndData(buf, metadata, data, offset, flags)
}

func PayloadOffset(b []byte) int {
	return header.FrameHeaderLength
}

func PayloadOffsetWithInitialN(b []byte) int {
	return initialNFieldOffset + sizeOfInt
}

func InitialRequestN(b []byte) uint32 {
	return header.Uint32(b, initialNFieldOffset)
}