package frame

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/keepalive"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/setup"
	"io"
)

// Frame is our single data container, it can take the shape of any
// frame type in the protocol, and is the structure we use to re-use
// data. While the struct itself contains some basic operators (like
// the ability to determine the type of the frame), most read operations
// are performed via the frame type-specific packages below this level,
// like `frame/setup`.
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

func (f *Frame) Flags() uint16 {
	return header.Flags(f.Buf)
}

func (f *Frame) Data() []byte {
	dataLength := f.dataLength()
	dataOffset := f.dataOffset()
	if 0 == dataLength {
		return nil
	}
	return f.Buf[dataOffset : dataOffset+dataLength]
}

func (f *Frame) Metadata() []byte {
	metadataLength := max(0, f.metadataFieldLength()-header.SizeOfInt)
	metadataOffset := f.payloadOffset() + header.SizeOfInt
	if 0 == metadataLength {
		return nil
	}
	return f.Buf[metadataOffset : metadataOffset+metadataLength]
}

// Make a copy of this frame. If target is provided, it will be
// used (potentially after resizing). If target is nil, a new
// frame will be allocated.
func (f *Frame) Copy(target *Frame) *Frame {
	if target == nil {
		target = &Frame{Buf: make([]byte, len(f.Buf))}
	}
	header.ResizeSlice(&target.Buf, len(f.Buf))
	copy(target.Buf, f.Buf)
	return target
}

// Get a human-readable description of this frame
func (f *Frame) Describe() string {
	switch f.Type() {
	case header.FTKeepAlive:
		return keepalive.Describe(f.Buf)
	case header.FTRequestN:
		return requestn.Describe(f.Buf)
	default:
		return fmt.Sprintf("UnknownFrame{type=%d, contents=% x}", f.Type(), f.Buf)
	}
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
	case header.FTRequestN:
		return requestn.PayloadOffset()
	}
	// TODO, I don't think we should use Panic like this; it's a legitimate
	//       condition that the remote implementation may send us invalid
	//       data, we should fail gracefully from that.
	panic(fmt.Sprintf("Unknown frame type: %d", f.Type()))
}

func (f *Frame) metadataFieldLength() int {
	if header.Flags(f.Buf)&header.FlagHasMetadata == 0 {
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
	sink               io.Writer
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
