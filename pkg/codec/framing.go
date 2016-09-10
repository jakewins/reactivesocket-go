package codec

import (
	"io"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/codec/setup"
	"fmt"
)

type FrameDecoder struct {
	source *io.Reader
}

func (d *FrameDecoder) Read(target *Frame) error {
	return nil
}

type FrameEncoder struct {
	sink *io.Writer
}

func (self *FrameEncoder) Write(frame *Frame) error {
	return nil
}

type Frame struct {
	Buf []byte
}

func (f *Frame) Type() uint16 {
	return header.FrameType(f.Buf)
}

func (f *Frame) Data() []byte {
	dataLength := f.dataLength()
	dataOffset := f.dataOffset()
	if 0 == dataLength {
		return nil
	}
	return f.Buf[dataOffset:dataOffset+dataLength]
}

func (f *Frame) Metadata() []byte {
	return nil
}

func (f *Frame) dataLength() int {
	frameLength := header.FrameLength(f.Buf)
	metadataLength := f.metadataFieldLength()
	return frameLength - metadataLength - f.payloadOffset()
}

// Return the byte offset where data starts in any given Frame
func (f *Frame) dataOffset() int {
	return f.payloadOffset() + f.metadataFieldLength()
}

func (f *Frame) payloadOffset() int {
	switch f.Type() {
	case header.FTSetup:
		return setup.PayloadOffset(f.Buf)
	}
	panic(fmt.Sprintf("Unknown frame type: %d", f.Type()))
}

func (f *Frame) metadataFieldLength() int {
	if header.Flags(f.Buf) & header.FlagHasMetadata == 0 {
		return 0
	}

	return int(header.Uint32(f.Buf, f.payloadOffset()))
}