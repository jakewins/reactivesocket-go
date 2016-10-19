package frame

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/cancel"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/errorc"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/keepalive"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/response"
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
	return header.Data(f.Buf, f.payloadOffset)
}
func (f *Frame) Metadata() []byte {
	return header.Metadata(f.Buf, f.payloadOffset)
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
	case header.FTResponse:
		return response.Describe(f.Buf)
	case header.FTFireAndForget:
		return request.Describe(f.Buf)
	case header.FTMetadataPush:
		return request.Describe(f.Buf)
	case header.FTRequestChannel:
		return request.Describe(f.Buf)
	case header.FTRequestStream:
		return request.Describe(f.Buf)
	case header.FTRequestSubscription:
		return request.Describe(f.Buf)
	case header.FTRequestResponse:
		return request.Describe(f.Buf)
	case header.FTError:
		return errorc.Describe(f.Buf)
	case header.FTCancel:
		return cancel.Describe(f.Buf)
	default:
		return fmt.Sprintf("UnknownFrame{type=%d, contents=% x}", f.Type(), f.Buf)
	}
}

func (f *Frame) payloadOffset() int {
	switch f.Type() {
	case header.FTSetup:
		return setup.PayloadOffset(f.Buf)
	case header.FTFireAndForget:
		return request.PayloadOffset()
	case header.FTMetadataPush:
		return request.PayloadOffset()
	case header.FTRequestResponse:
		return request.PayloadOffset()
	case header.FTRequestChannel:
		return request.PayloadOffsetWithInitialN()
	case header.FTRequestStream:
		return request.PayloadOffsetWithInitialN()
	case header.FTRequestSubscription:
		return request.PayloadOffsetWithInitialN()
	case header.FTRequestN:
		return requestn.PayloadOffset()
	case header.FTResponse:
		return response.PayloadOffset()
	case header.FTError:
		return errorc.PayloadOffset()
	case header.FTCancel:
		return cancel.PayloadOffset()
	}
	// TODO, I don't think we should use Panic like this; it's a legitimate
	//       condition that the remote implementation may send us invalid
	//       data, we should fail gracefully from that.
	panic(fmt.Sprintf("Unknown frame type: %d", f.Type()))
}

func (f *Frame) metadataFieldLength() int {
	return header.MetadataFieldLength(f.Buf, f.payloadOffset)
}

// Message creation
// Note that these only create messages, reading message-specific fields
// is done via the accessor methods under `frame/<msgtype>#..(*Frame)`
func EncodeResponse(f *Frame, streamId uint32, flags uint16, metadata, data []byte) *Frame {
	response.Encode(&f.Buf, streamId, flags, metadata, data)
	return f
}
func EncodeSetup(f *Frame, flags uint16, keepaliveInterval, maxLifetime uint32,
	metadataMimeType, dataMimeType string, metadata, data []byte) *Frame {
	setup.Encode(&f.Buf, flags, keepaliveInterval, maxLifetime, metadataMimeType, dataMimeType, metadata, data)
	return f
}
func EncodeKeepalive(f *Frame, respond bool) *Frame {
	keepalive.Encode(&f.Buf, respond)
	return f
}
func EncodeRequest(f *Frame, streamId uint32, flags, frameType uint16, metadata, data []byte) *Frame {
	request.Encode(&f.Buf, streamId, flags, frameType, metadata, data)
	return f
}
func EncodeRequestWithInitialN(f *Frame, streamId, initialN uint32, flags, frameType uint16, metadata, data []byte) *Frame {
	request.EncodeWithInitialN(&f.Buf, streamId, initialN, flags, frameType, metadata, data)
	return f
}
func EncodeRequestN(f *Frame, streamId, requestN uint32) *Frame {
	requestn.Encode(&f.Buf, streamId, requestN)
	return f
}
func EncodeCancel(f *Frame, streamId uint32) *Frame {
	cancel.Encode(&f.Buf, streamId)
	return f
}
func EncodeError(f *Frame, streamId, errorCode uint32, metadata, data []byte) *Frame {
	errorc.Encode(&f.Buf, streamId, errorCode, metadata, data)
	return f
}

func Setup(flags uint16, keepaliveInterval, maxLifetime uint32,
	metadataMimeType, dataMimeType string, metadata, data []byte) *Frame {
	return EncodeSetup(&Frame{}, flags, keepaliveInterval, maxLifetime, metadataMimeType, dataMimeType, metadata, data)
}
func Response(streamId uint32, flags uint16, metadata, data []byte) *Frame {
	return EncodeResponse(&Frame{}, streamId, flags, metadata, data)
}
func Keepalive(respond bool) *Frame {
	return EncodeKeepalive(&Frame{}, respond)
}
func Request(streamId uint32, flags, frameType uint16, metadata, data []byte) *Frame {
	return EncodeRequest(&Frame{}, streamId, flags, frameType, metadata, data)
}
func RequestWithInitialN(streamId, initialN uint32, flags, frameType uint16, metadata, data []byte) *Frame {
	return EncodeRequestWithInitialN(&Frame{}, streamId, initialN, flags, frameType, metadata, data)
}
func RequestN(streamId, requestN uint32) *Frame {
	return EncodeRequestN(&Frame{}, streamId, requestN)
}
func Error(streamId, errorCode uint32, metadata, data []byte) *Frame {
	return EncodeError(&Frame{}, streamId, errorCode, metadata, data)
}
func Cancel(streamId uint32) *Frame {
	return EncodeCancel(&Frame{}, streamId)
}

// Frame encoder/decoder below should be moved out of here

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
	frameLength := uint32(len(frame.Buf) + header.SizeOfInt)
	header.PutUint32(e.frameLengthScratch, 0, frameLength)

	_, err := e.sink.Write(e.frameLengthScratch)
	return err
}
