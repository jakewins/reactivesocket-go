package header

import (
"encoding/binary"
)

// Common for all frames

const (
	SizeOfInt = 4
	SizeOfShort = 2
	frameLengthFieldOffset = 0
	typeFieldOffset = frameLengthFieldOffset + SizeOfInt
	flagsFieldOffset = typeFieldOffset + SizeOfShort
	streamIdFieldOffset = flagsFieldOffset + SizeOfShort
	payloadOffset = streamIdFieldOffset + SizeOfInt
	FrameHeaderLength = payloadOffset
)

const (
	FlagHasMetadata uint16 = 1 << 15
)

const (
	FTSetup uint16 = 0x01
)

func computeMetadataLength(metadataPayloadLength int) int {
	if metadataPayloadLength == 0 {
		return 0
	}
	return metadataPayloadLength + SizeOfInt
}

func computeFrameHeaderLength(metadataLength, dataLength int) int {
	return payloadOffset + computeMetadataLength(metadataLength) + dataLength
}

func ComputeFrameLength(metadataMimeType, dataMimeType string, metadata, data []byte) uint32 {
	length := computeFrameHeaderLength(len(metadata), len(data))
	length += SizeOfInt * 3
	length += 1 + len(metadataMimeType)
	length += 1 + len(dataMimeType)
	return uint32(length)
}

func PutUint16(b []byte, offset int, v uint16) {
	binary.BigEndian.PutUint16(b[offset:], v)
}

func PutUint32(b []byte, offset int, v uint32) {
	binary.BigEndian.PutUint32(b[offset:], v)
}

func Uint16(b []byte, offset int) uint16 {
	return binary.BigEndian.Uint16(b[offset:])
}

func Uint32(b []byte, offset int) uint32 {
	return binary.BigEndian.Uint32(b[offset:])
}

func PutMimeType(b []byte, offset int, mimeType string) int {
	length := len(mimeType)
	b[offset] = byte(length)
	copy(b[offset+1:offset+1+length], mimeType)
	return 1 + length
}

func MimeType(b []byte, offset int) string {
	length := int(b[offset])
	return string(b[offset+1:offset+1+length])
}

func EncodeHeader(buf []byte, frameLength uint32, flags uint16, ft uint16, streamId uint32) {
	PutUint32(buf, frameLengthFieldOffset, frameLength)
	PutUint16(buf, typeFieldOffset, ft)
	PutUint16(buf, flagsFieldOffset, flags)
	PutUint32(buf, streamIdFieldOffset, streamId)
}


func Flags(b []byte) uint16 {
	return Uint16(b, flagsFieldOffset)
}

func FrameLength(b []byte) int {
	return int(Uint32(b, frameLengthFieldOffset))
}

func FrameType(b []byte) uint16 {
	return Uint16(b, typeFieldOffset)
}