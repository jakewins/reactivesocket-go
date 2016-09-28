package request

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
)

const (
	RequestFlagRequestChannelC = 1 << 12
	RequestFlagRequestChannelN = 1 << 11
)

const (
	sizeOfInt           = header.SizeOfInt
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

func PayloadOffset() int {
	return header.FrameHeaderLength
}

func PayloadOffsetWithInitialN() int {
	return initialNFieldOffset + sizeOfInt
}

func InitialRequestN(b []byte) uint32 {
	switch header.FrameType(b) {
	case header.FTFireAndForget:
		return 0
	case header.FTRequestResponse:
		return 1
	case header.FTRequestChannel:
		return header.Uint32(b, initialNFieldOffset)
	case header.FTRequestStream:
		return header.Uint32(b, initialNFieldOffset)
	case header.FTRequestSubscription:
		return header.Uint32(b, initialNFieldOffset)
	}
	panic(fmt.Sprintf("Expected a request frame, got %d", header.FrameType(b)))
}

func Describe(b []byte) string {
	var frameName string
	switch header.FrameType(b) {
	case header.FTRequestChannel:
		frameName = "RequestChannel"
	case header.FTRequestSubscription:
		frameName = "RequestSubscription"
	case header.FTRequestStream:
		frameName = "RequestStream"
	case header.FTRequestResponse:
		frameName = "RequestResponse"
	default:
		panic(fmt.Sprintf("Expected a request frame, got %d", header.FrameType(b)))
	}
	return fmt.Sprintf("%s{streamId=%d, initialRequestN=%d}",
		frameName, header.StreamID(b), InitialRequestN(b))
}
