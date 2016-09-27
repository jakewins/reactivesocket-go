package header

import (
	"encoding/binary"
)

// Common for all frames

const (
	SizeOfInt           = 4
	SizeOfShort         = 2
	typeFieldOffset     = 0
	flagsFieldOffset    = typeFieldOffset + SizeOfShort
	streamIdFieldOffset = flagsFieldOffset + SizeOfShort
	payloadOffset       = streamIdFieldOffset + SizeOfInt
	FrameHeaderLength   = payloadOffset
)

const (
	FlagHasMetadata      uint16 = 1 << 15
	FlagKeepaliveRespond        = 1 << 13
)

const (
	// Connection
	FTSetup     uint16 = 0x01
	FTLease            = 0x02
	FTKeepAlive        = 0x03
	// Requester to start request
	FTRequestResponse     = 0x04
	FTFireAndForget       = 0x05
	FTRequestStream       = 0x06
	FTRequestSubscription = 0x07
	FTRequestChannel      = 0x08
	// Requester mid-stream
	FTRequestN = 0x09
	FTCancel   = 0x0A
	// Responder
	FTResponse = 0x0B
	FTError    = 0x0C
	// Requester & Responder
	FTMetadataPush = 0x0D
	// synthetic types from Responder for use by the rest of the machinery
	FTNext         = 0x0E
	FTComplete     = 0x0F
	FTNextComplete = 0x10
)

func computeMetadataLength(metadataPayloadLength int) int {
	if metadataPayloadLength == 0 {
		return 0
	}
	return metadataPayloadLength + SizeOfInt
}

func ComputeLength(metadataLength, dataLength int) int {
	return payloadOffset + computeMetadataLength(metadataLength) + dataLength
}

func PutMimeType(b []byte, offset int, mimeType string) int {
	length := len(mimeType)
	b[offset] = byte(length)
	copy(b[offset+1:offset+1+length], mimeType)
	return 1 + length
}

func MimeType(b []byte, offset int) string {
	length := int(b[offset])
	return string(b[offset+1 : offset+1+length])
}

func EncodeHeader(buf []byte, flags uint16, ft uint16, streamId uint32) {
	PutUint16(buf, typeFieldOffset, ft)
	PutUint16(buf, flagsFieldOffset, flags)
	PutUint32(buf, streamIdFieldOffset, streamId)
}

func Flags(b []byte) uint16 {
	return Uint16(b, flagsFieldOffset)
}

func FrameType(b []byte) uint16 {
	return Uint16(b, typeFieldOffset)
}
func StreamID(b []byte) uint32 {
	return Uint32(b, streamIdFieldOffset)
}
func DataOffset(buf []byte, payloadOffset func() int) int {
	return payloadOffset() + MetadataFieldLength(buf, payloadOffset)
}
func MetadataFieldLength(buf []byte, payloadOffset func() int) int {
	if Flags(buf)&FlagHasMetadata == 0 {
		return 0
	}

	return int(Uint32(buf, payloadOffset()))
}
func Metadata(buf []byte, payloadOffset func() int) []byte {
	metadataLength := max(0, MetadataFieldLength(buf, payloadOffset)-SizeOfInt)
	if 0 == metadataLength {
		return nil
	}
	metadataOffset := payloadOffset() + SizeOfInt
	return buf[metadataOffset : metadataOffset+metadataLength]
}
func Data(buf []byte, payloadOffset func() int) []byte {
	dataLength := dataLength(buf, payloadOffset)
	if 0 == dataLength {
		return nil
	}

	dataOffset := DataOffset(buf, payloadOffset)
	return buf[dataOffset : dataOffset+dataLength]
}

func dataLength(buf []byte, payloadOffset func() int) int {
	frameLength := len(buf)
	metadataLength := MetadataFieldLength(buf, payloadOffset)
	return frameLength - metadataLength - payloadOffset()
}

func EncodeMetaDataAndData(buf, metadata, data []byte, offset int, flags uint16) {
	if flags&FlagHasMetadata != 0 {
		PutUint32(buf, offset, uint32(len(metadata)+SizeOfInt))
		offset += SizeOfInt
		copy(buf[offset:offset+len(metadata)], metadata)
		offset += len(metadata)
	}

	if len(data) > 0 {
		copy(buf[offset:], data)
	}
}

// Below are general-ish utility methods that are used by the codec packages

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

// Ensure the given pointer refers to a slice with at least the specified capacity,
// allocating a new underlying array for the slice to point to if not
// Returns the resized slice.
// TODO: I really don't like this, it's a result of Codec not depending on Frame,
//       and using len(frame.Buf) to track frame length. It'd be nicer to have someting
//       that took a regular slice and returned another, like append() does.
func ResizeSlice(slicePtr *[]byte, ensure int) []byte {
	slice := *slicePtr
	if ensure > cap(slice) {
		remainder := ensure % 512
		if remainder == 0 {
			*slicePtr = make([]byte, ensure)
		} else {
			*slicePtr = make([]byte, ensure+(512-remainder))
		}
	}

	// Replace the slice struct with one that's limited to the set length
	*slicePtr = (*slicePtr)[:ensure]
	return *slicePtr
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
