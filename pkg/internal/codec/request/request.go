package request

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
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
	flags |= header.FlagRequestChannelInitialN

	header.EncodeHeader(buf, flags, frameType, streamId)

	header.PutUint32(buf, initialNFieldOffset, initialN)

	offset := initialNFieldOffset
	offset += sizeOfInt

	header.EncodeMetaDataAndData(buf, metadata, data, offset, flags)
}

func PayloadOffset(b []byte) int {
	if header.Flags(b)&header.FlagRequestChannelInitialN == 0 {
		return header.FrameHeaderLength
	}
	return header.FrameHeaderLength + sizeOfInt
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

	if header.Flags(b)&header.FlagRequestChannelInitialN == 0 {
		return 0
	}
	return header.Uint32(b, initialNFieldOffset)
}

func Describe(b []byte) string {
	var frameName string
	var payloadOffset = PayloadOffset(b)

	switch header.FrameType(b) {
	case header.FTRequestChannel:
		frameName = "RequestChannel"
		if header.Flags(b)&header.FlagRequestChannelComplete != 0 {
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

	var initialN = ""
	if header.Flags(b)&header.FlagRequestChannelInitialN != 0 {
		initialN = fmt.Sprintf(" initialRequestN=%d,", InitialRequestN(b))
	}

	return fmt.Sprintf("%s{streamId=%d,%s metadata=[% x], data=[% x]}",
		frameName, header.StreamID(b), initialN,
		header.Metadata(b, payloadOffset), header.Data(b, payloadOffset))
}
