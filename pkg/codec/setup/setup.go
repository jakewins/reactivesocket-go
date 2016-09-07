package setup

import (
	"github.com/jakewins/reactivesocket-go/pkg/codec"
)

// Common for all frames

const (
	sizeOfInt = 4
	sizeOfShort = 2
	frameLengthFieldOffset = 0
	typeFieldOffset = frameLengthFieldOffset + sizeOfInt
	flagsFieldOffset = typeFieldOffset + sizeOfShort
	streamIdFieldOffset = flagsFieldOffset + sizeOfShort
	payloadOffset = streamIdFieldOffset + sizeOfInt
	frameHeaderLength = payloadOffset
)


const (
	FlagHasMetadata uint16 = 1 << 15
)

// Setup frame specific

const (
	currentVersion = 0
	versionFieldOffset = frameHeaderLength
	keepaliveIntervalFieldOffset = versionFieldOffset + sizeOfInt
	maxLifetimeFieldOffset = keepaliveIntervalFieldOffset + sizeOfInt
	metadataMimeTypeLengthOffset = maxLifetimeFieldOffset + sizeOfInt
)

const (
	FTSetup uint16 = 0x01
)

func computeMetadataLength(metadataPayloadLength int) int {
	if metadataPayloadLength == 0 {
		return 0
	}
	return metadataPayloadLength + sizeOfInt
}

func computeFrameHeaderLength(metadataLength, dataLength int) int {
	return payloadOffset + computeMetadataLength(metadataLength) + dataLength
}

func computeFrameLength(metadataMimeType, dataMimeType string, metadata, data []byte) uint32 {
	length := computeFrameHeaderLength(len(metadata), len(data))
	length += sizeOfInt * 3
	length += 1 + len(metadataMimeType)
	length += 1 + len(dataMimeType)
	return uint32(length)
}

func putUint16(b []byte, offset int, v uint16) {
	b[offset+0] = byte(v >> 8)
	b[offset+1] = byte(v)
}

func putUint32(b []byte, offset int, v uint32) {
	b[offset+0] = byte(v >> 24)
	b[offset+1] = byte(v >> 16)
	b[offset+2] = byte(v >> 8)
	b[offset+3] = byte(v)
}

func putMimeType(b []byte, offset int, mimeType string) int {
	length := len(mimeType)
	b[offset] = byte(length)
	copy(b[offset+1:offset+1+length], mimeType)
	return 1 + length
}

func encodeFrameHeader(buf []byte, frameLength uint32, flags uint16, ft uint16, streamId uint32) {
	putUint32(buf, frameLengthFieldOffset, frameLength)
	putUint16(buf, typeFieldOffset, ft)
	putUint16(buf, flagsFieldOffset, flags)
	putUint32(buf, streamIdFieldOffset, streamId)
}

func NewSetupFrame(flags uint16, keepaliveInterval, maxLifetime uint32,
                   metadataMimeType, dataMimeType string,
                   metadata, data []byte) *codec.Frame {
	frameLength := computeFrameLength(metadataMimeType, dataMimeType, metadata, data)
	buf := make([]byte, frameLength)

	if len(metadata) > 0 {
		flags |= FlagHasMetadata
	}

	encodeFrameHeader(buf, frameLength, flags, FTSetup, 0)
	putUint32(buf, versionFieldOffset, currentVersion)
	putUint32(buf, keepaliveIntervalFieldOffset, keepaliveInterval)
	putUint32(buf, maxLifetimeFieldOffset, maxLifetime)

	offset := frameHeaderLength
	offset += sizeOfInt * 3 // The three ints we write above
	offset += putMimeType(buf, offset, metadataMimeType)
	offset += putMimeType(buf, offset, dataMimeType)

	if flags & FlagHasMetadata != 0 {
		// TODO: Why are we adding sizeOfInt here; this follows what the Java code does, and I'm sure it's fine and well,
		// but I want to understand why we include this size here, isn't it redundant? Need to check the spec
		putUint32(buf, offset, uint32(len(metadata) + sizeOfInt))
		offset += sizeOfInt
		copy(buf[offset:offset+len(metadata)], metadata)
		offset += len(metadata)
	}

	if len(data) > 0 {
		copy(buf[offset:], data)
	}

	return &codec.Frame{Buf:buf}
}

func Flags(f *codec.Frame) uint16 {
	return 0
}

func Version(f *codec.Frame) uint {
	return 0
}

func KeepaliveInterval(f *codec.Frame) uint {
	return 0
}

func MaxLifetime(f *codec.Frame) uint {
	return 0
}

func MetadataMimeType(f *codec.Frame) string {
	return ""
}

func DataMimeType(f *codec.Frame) string {
	return ""
}

