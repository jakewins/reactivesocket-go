package setup

import "github.com/jakewins/reactivesocket-go/pkg/codec"

// Common for all frames

const (
	sizeOfInt = 4
	typeFieldOffset = sizeOfInt
	flagsFieldOffset = typeFieldOffset + sizeOfInt
	streamIdFieldOffset = flagsFieldOffset + sizeOfInt
	payloadOffset = streamIdFieldOffset + sizeOfInt
	frameHeaderLength = payloadOffset
)

// Setup frame specific

const (
	versionFieldOffset = frameHeaderLength
	keepaliveIntervalFieldOffset = versionFieldOffset + sizeOfInt
	maxLifetimeFieldOffset = keepaliveIntervalFieldOffset + sizeOfInt
	metadataMimeTypeLengthOffset = maxLifetimeFieldOffset + sizeOfInt
)

func computeMetadataLength(metadataPayloadLength uint) uint {
	if metadataPayloadLength == 0 {
		return 0
	}
	return metadataPayloadLength + sizeOfInt
}

func computeFrameHeaderLength(metadataLength, dataLength uint) uint {
	return payloadOffset + computeMetadataLength(metadataLength) + dataLength
}

func computeFrameLength(metadataMimeType, dataMimeType string, metadata, data []byte) uint {
	length := computeFrameHeaderLength(len(metadata), len(data))
	length += sizeOfInt * 3
	length += 1 + len(metadataMimeType)
	length += 1 + len(dataMimeType)
	return length
}


func encodeFrameHeader(buf []byte, frameLength, flags, streamId uint,  )

func NewSetupFrame(flags, keepaliveInterval, maxLifetime uint,
                   metadataMimeType, dataMimeType string,
                   metadata, data []byte) codec.Frame {
	frameLength := computeFrameLength(metadataMimeType, dataMimeType, metadata, data)
	buf := make([]byte, frameLength)



	return codec.Frame{buf}
}

func Flags(f *codec.Frame) uint {
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
	return nil
}

func DataMimeType(f *codec.Frame) string {
	return nil
}

