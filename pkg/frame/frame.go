package frame


import (
	"io"
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/codec/setup"
)

type Frame struct {
	// This is a slice that is sized to fit the current frame
	Buf []byte
}

func (f *Frame) Type() uint16 {
	return header.FrameType(f.Buf)
}

func (f *Frame) StreamID() uint32 {
	return header.StreamID(f.Buf)
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
	metadataLength := max(0, f.metadataFieldLength() - header.SizeOfInt)
	metadataOffset := f.payloadOffset() + header.SizeOfInt
	if 0 == metadataLength {
		return nil
	}
	return f.Buf[metadataOffset:metadataOffset+metadataLength]
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

type FrameDecoder struct {
	Source io.Reader
}

func (d *FrameDecoder) Read(target *Frame) error {
	frameLength, err := d.readFrameLength(target)
	if err != nil {
		return err
	}
	header.ResizeSlice(&target.Buf, frameLength)

	restOfFrameSlice := target.Buf[header.SizeOfInt:frameLength]
	_, err = io.ReadFull(d.Source, restOfFrameSlice)
	if err != nil {
		return err
	}

	return nil
}

func (d *FrameDecoder) readFrameLength(target *Frame) (int, error) {
	header.ResizeSlice(&target.Buf, header.SizeOfInt)
	frameSizeSlice := target.Buf[:header.SizeOfInt]

	_, err := io.ReadFull(d.Source, frameSizeSlice)
	if err != nil {
		return 0, err
	}

	return header.FrameLength(target.Buf), nil
}

type FrameEncoder struct {
	Sink io.Writer
}

func (self *FrameEncoder) Write(frame *Frame) error {
	_, err := self.Sink.Write(frame.Buf)
	return err
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
