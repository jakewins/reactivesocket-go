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
	if initialN == 0 {
		Encode(bufPtr, streamId, flags, frameType, metadata, data)
		return
	}

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
	case header.FTMetadataPush:
		return 0
	case header.FTFireAndForget:
		return 0
	case header.FTRequestResponse:
		return 1
	// These cases just verify that the frame is a Request frame
	case header.FTRequestChannel:
	case header.FTRequestStream:
	case header.FTRequestSubscription:
	default:
		panic(fmt.Sprintf("Expected a request frame, got %d", header.FrameType(b)))
	}

	if header.Flags(b)&RequestFlagRequestChannelN == 0 {
		return 0
	}
	return header.Uint32(b, initialNFieldOffset)
}

func Describe(b []byte) string {
	var frameName string
	var payloadOffset = PayloadOffset
	if header.Flags(b)&RequestFlagRequestChannelN != 0 {
		payloadOffset = PayloadOffsetWithInitialN
	}

	switch header.FrameType(b) {
	case header.FTRequestChannel:
		frameName = "RequestChannel"
		if header.Flags(b)&RequestFlagRequestChannelC != 0 {
			frameName = "RequestChannel[Complete]"
		}
	case header.FTRequestSubscription:
		frameName = "RequestSubscription"
	case header.FTRequestStream:
		frameName = "RequestStream"
	case header.FTRequestResponse:
		frameName = "RequestResponse"
	case header.FTFireAndForget:
		frameName = "FireAndForget"
	case header.FTMetadataPush:
		frameName = "MetadataPush"
	default:
		panic(fmt.Sprintf("Expected a request frame, got %d", header.FrameType(b)))
	}

	return fmt.Sprintf("%s{streamId=%d, initialRequestN=%d, metadata=[% x], data=[% x]}",
		frameName, header.StreamID(b), InitialRequestN(b),
		header.Metadata(b, payloadOffset), header.Data(b, payloadOffset))
}
