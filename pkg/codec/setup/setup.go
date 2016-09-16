package setup

import (
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
)

const (
	SetupFlagWillHonorLease = 1 << 13
	SetupFlagStrictInterpretation = 1 << 12
)

const (
	sizeOfInt = header.SizeOfInt
	currentVersion = 0
	versionFieldOffset = header.FrameHeaderLength
	keepaliveIntervalFieldOffset = versionFieldOffset + sizeOfInt
	maxLifetimeFieldOffset = keepaliveIntervalFieldOffset + sizeOfInt
	metadataMimeTypeLengthOffset = maxLifetimeFieldOffset + sizeOfInt
)


func computeFrameLength(metadataMimeType, dataMimeType string, metadata, data []byte) int {
	length := header.ComputeLength(len(metadata), len(data))
	length += sizeOfInt * 3
	length += 1 + len(metadataMimeType)
	length += 1 + len(dataMimeType)
	return length
}

func Encode(bufPtr *[]byte, flags uint16, keepaliveInterval, maxLifetime uint32,
						metadataMimeType, dataMimeType string,
						metadata, data []byte) {
	frameLength := computeFrameLength(metadataMimeType, dataMimeType, metadata, data)

	header.EnsureCapacity(bufPtr, frameLength)
	buf := *bufPtr

	if len(metadata) > 0 {
		flags |= header.FlagHasMetadata
	}

	header.EncodeHeader(buf, uint32(frameLength), flags, header.FTSetup, 0)
	header.PutUint32(buf, versionFieldOffset, currentVersion)
	header.PutUint32(buf, keepaliveIntervalFieldOffset, keepaliveInterval)
	header.PutUint32(buf, maxLifetimeFieldOffset, maxLifetime)

	offset := header.FrameHeaderLength
	offset += sizeOfInt * 3 // The three ints we write above
	offset += header.PutMimeType(buf, offset, metadataMimeType)
	offset += header.PutMimeType(buf, offset, dataMimeType)

	if flags & header.FlagHasMetadata != 0 {
		header.PutUint32(buf, offset, uint32(len(metadata) + sizeOfInt))
		offset += sizeOfInt
		copy(buf[offset:offset+len(metadata)], metadata)
		offset += len(metadata)
	}

	if len(data) > 0 {
		copy(buf[offset:], data)
	}
}

func Flags(b []byte) uint16 {
	return header.Flags(b) & (SetupFlagWillHonorLease | SetupFlagStrictInterpretation)
}

func Version(b []byte) uint32 {
	return header.Uint32(b, versionFieldOffset)
}

func KeepaliveInterval(b []byte) uint32 {
	return header.Uint32(b, keepaliveIntervalFieldOffset)
}

func MaxLifetime(b []byte) uint32 {
	return header.Uint32(b, maxLifetimeFieldOffset)
}

func MetadataMimeType(b []byte) string {
	return header.MimeType(b, metadataMimeTypeLengthOffset)
}

func DataMimeType(b []byte) string {
	offset := int(metadataMimeTypeLengthOffset)
	offset += 1 + int(b[offset])
	return header.MimeType(b, offset)
}

func PayloadOffset(b []byte) int {
	offset := metadataMimeTypeLengthOffset

	metadataMimeTypeLength := int(b[offset])
	offset += 1 + metadataMimeTypeLength

	dataMimeTypeLength := int(b[offset])
	offset += 1 + dataMimeTypeLength

	return offset
}

