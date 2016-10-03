package errorc // because 'error' clashes with builtin error

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
)

const (
	sizeOfInt            = header.SizeOfInt
	errorCodeFieldOffset = header.FrameHeaderLength
)

const (
	ECInvalidSetup     uint32 = 0x0001
	ECUnsupportedSetup        = 0x0002
	ECRejectedSetup           = 0x0003
	ECConnectionError         = 0x0101
	ECApplicationError        = 0x0201
	ECRejected                = 0x0202
	ECCancel                  = 0x0203
	ECInvalid                 = 0x0204
)

func computeFrameLength(metadata, data []byte) int {
	length := header.ComputeLength(len(metadata), len(data))
	length += sizeOfInt
	return length
}

func Encode(bufPtr *[]byte, streamId, errorCode uint32, metadata, data []byte) {
	buf := header.ResizeSlice(bufPtr, computeFrameLength(metadata, data))
	var flags uint16 = 0
	if len(metadata) > 0 {
		flags |= header.FlagHasMetadata
	}

	header.EncodeHeader(buf, flags, header.FTError, streamId)
	header.PutUint32(buf, errorCodeFieldOffset, errorCode)

	offset := errorCodeFieldOffset + sizeOfInt
	header.EncodeMetaDataAndData(buf, metadata, data, offset, flags)
}

func PayloadOffset() int {
	return errorCodeFieldOffset + sizeOfInt
}

func ErrorCode(buf []byte) uint32 {
	return header.Uint32(buf, errorCodeFieldOffset)
}

func Describe(buf []byte) string {
	return fmt.Sprintf("Error{streamId=%d, errorCode=%d, metadata=[% x], data=[% x]}",
		header.StreamID(buf), ErrorCode(buf), header.Metadata(buf, PayloadOffset), header.Data(buf, PayloadOffset))
}
