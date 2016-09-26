package frame


import (
	"io"
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/setup"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/request"
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
	frameLength := len(f.Buf)
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
	case header.FTFireAndForget:
		return request.PayloadOffset(f.Buf)
	case header.FTRequestResponse:
		return request.PayloadOffset(f.Buf)
	case header.FTRequestChannel:
		return request.PayloadOffsetWithInitialN(f.Buf)
	case header.FTRequestStream:
		return request.PayloadOffsetWithInitialN(f.Buf)
	case header.FTRequestSubscription:
		return request.PayloadOffsetWithInitialN(f.Buf)
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
	source io.Reader
}

func (d *FrameDecoder) Read(target *Frame) error {
	frameLength, err := d.readFrameLength(target)
	if err != nil {
		return err
	}
	header.ResizeSlice(&target.Buf, frameLength)

	_, err = io.ReadFull(d.source, target.Buf)
	if err != nil {
		return err
	}

	return nil
}

func (d *FrameDecoder) readFrameLength(target *Frame) (int, error) {
	header.ResizeSlice(&target.Buf, header.SizeOfInt)
	frameSizeSlice := target.Buf[:header.SizeOfInt]

	_, err := io.ReadFull(d.source, frameSizeSlice)
	if err != nil {
		return 0, err
	}

	frameLength := int(header.Uint32(target.Buf, 0) - header.SizeOfInt)
	return frameLength, nil
}

func NewFrameDecoder(source io.Reader) *FrameDecoder {
	return &FrameDecoder{source}
}

type FrameEncoder struct {
	sink io.Writer
	frameLengthScratch []byte
}

func NewFrameEncoder(sink io.Writer) *FrameEncoder {
	return &FrameEncoder{sink, make([]byte, 4)}
}

func (e *FrameEncoder) Write(frame *Frame) error {
	if err := e.writeFrameLength(frame); err != nil {
		return err
	}
	_, err := e.sink.Write(frame.Buf)
	return err
}

func (e *FrameEncoder) writeFrameLength(frame *Frame) error {
	if len(e.frameLengthScratch) != 4 {
		e.frameLengthScratch = make([]byte, 4)
	}
	frameLength := uint32(len(frame.Buf) + header.SizeOfInt)
	header.PutUint32(e.frameLengthScratch, 0, frameLength)

	_, err := e.sink.Write(e.frameLengthScratch)
	return err
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
