package request_test

import (
	"bytes"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/request"
	"testing"
)

func TestFireAndForgetFrameEncoding(t *testing.T) {
	var flags uint16 = 0
	metadata := []byte{1, 2, 3}
	data := []byte{4, 5, 6}
	var streamId uint32 = 1337

	f := frame.Request(streamId, flags, header.FTFireAndForget, metadata, data)

	if request.InitialRequestN(f) != 0 {
		t.Errorf("Expected initial request N to be %d, found %s", 0, request.InitialRequestN(f))
	}
	if f.Type() != header.FTFireAndForget {
		t.Errorf("Expected type to be %d, found %d", header.FTRequestChannel, f.Type())
	}
	if f.StreamID() != streamId {
		t.Errorf("Expected stream id to be %d, found %d", streamId, f.StreamID())
	}
	if !bytes.Equal(f.Data(), data) {
		t.Errorf("Expected frame data to be `% x` but found `% x`", data, f.Data())
	}
	if !bytes.Equal(f.Metadata(), metadata) {
		t.Errorf("Expected frame metadata to be `% x` but found `% x`", metadata, f.Metadata())
	}
	if len(f.Buf) != 18 {
		t.Errorf("Expected frame length to be %d but found %d", 18, len(f.Buf))
	}
}

func TestRequestResponseFrameEncoding(t *testing.T) {
	var flags uint16 = 0
	metadata := []byte{1, 2, 3}
	data := []byte{4, 5, 6}
	var streamId uint32 = 1337

	f := frame.Request(streamId, flags, header.FTRequestResponse, metadata, data)

	if request.InitialRequestN(f) != 1 {
		t.Errorf("Expected initial request N to be %d, found %s", 1, request.InitialRequestN(f))
	}
	if f.Type() != header.FTRequestResponse {
		t.Errorf("Expected type to be %d, found %d", header.FTRequestResponse, f.Type())
	}
	if f.StreamID() != streamId {
		t.Errorf("Expected stream id to be %d, found %d", streamId, f.StreamID())
	}
	if !bytes.Equal(f.Data(), data) {
		t.Errorf("Expected frame data to be `% x` but found `% x`", data, f.Data())
	}
	if !bytes.Equal(f.Metadata(), metadata) {
		t.Errorf("Expected frame metadata to be `% x` but found `% x`", metadata, f.Metadata())
	}
	if len(f.Buf) != 18 {
		t.Errorf("Expected frame length to be %d but found %d", 18, len(f.Buf))
	}
}

func TestRequestChannelFrameEncoding(t *testing.T) {
	var flags uint16 = 0
	metadata := []byte{1, 2, 3}
	data := []byte{4, 5, 6}
	var streamId uint32 = 7331
	var initialRequestN uint32 = 1338

	f := frame.RequestWithInitialN(streamId, initialRequestN, flags, header.FTRequestChannel, metadata, data)

	if request.InitialRequestN(f) != initialRequestN {
		t.Errorf("Expected initial request N to be %d, found %d", initialRequestN, request.InitialRequestN(f))
	}
	if f.Type() != header.FTRequestChannel {
		t.Errorf("Expected type to be %d, found %d", header.FTRequestResponse, f.Type())
	}
	if f.StreamID() != streamId {
		t.Errorf("Expected stream id to be %d, found %d", streamId, f.StreamID())
	}
	if !bytes.Equal(f.Data(), data) {
		t.Errorf("Expected frame data to be `% x` but found `% x`", data, f.Data())
	}
	if !bytes.Equal(f.Metadata(), metadata) {
		t.Errorf("Expected frame metadata to be `% x` but found `% x`", metadata, f.Metadata())
	}
	if len(f.Buf) != 22 {
		t.Errorf("Expected frame length to be %d but found %d", 22, len(f.Buf))
	}
}

func TestRequestStreamFrameEncoding(t *testing.T) {
	var flags uint16 = 0
	metadata := []byte{1, 2, 3}
	data := []byte{4, 5, 6}
	var streamId uint32 = 7331
	var initialRequestN uint32 = 1338

	f := frame.RequestWithInitialN(streamId, initialRequestN, flags, header.FTRequestStream, metadata, data)

	if request.InitialRequestN(f) != initialRequestN {
		t.Errorf("Expected initial request N to be %d, found %d", initialRequestN, request.InitialRequestN(f))
	}
	if f.Type() != header.FTRequestStream {
		t.Errorf("Expected type to be %d, found %d", header.FTRequestStream, f.Type())
	}
	if f.StreamID() != streamId {
		t.Errorf("Expected stream id to be %d, found %d", streamId, f.StreamID())
	}
	if !bytes.Equal(f.Data(), data) {
		t.Errorf("Expected frame data to be `% x` but found `% x`", data, f.Data())
	}
	if !bytes.Equal(f.Metadata(), metadata) {
		t.Errorf("Expected frame metadata to be `% x` but found `% x`", metadata, f.Metadata())
	}
	if len(f.Buf) != 22 {
		t.Errorf("Expected frame length to be %d but found %d", 22, len(f.Buf))
	}
}

func TestRequestSubscriptionFrameEncoding(t *testing.T) {
	var flags uint16 = 0
	metadata := []byte{1, 2, 3}
	data := []byte{4, 5, 6}
	var streamId uint32 = 7331
	var initialRequestN uint32 = 1338

	f := frame.RequestWithInitialN(streamId, initialRequestN, flags, header.FTRequestSubscription, metadata, data)

	if request.InitialRequestN(f) != initialRequestN {
		t.Errorf("Expected initial request N to be %d, found %d", initialRequestN, request.InitialRequestN(f))
	}
	if f.Type() != header.FTRequestSubscription {
		t.Errorf("Expected type to be %d, found %d", header.FTRequestSubscription, f.Type())
	}
	if f.StreamID() != streamId {
		t.Errorf("Expected stream id to be %d, found %d", streamId, f.StreamID())
	}
	if !bytes.Equal(f.Data(), data) {
		t.Errorf("Expected frame data to be `% x` but found `% x`", data, f.Data())
	}
	if !bytes.Equal(f.Metadata(), metadata) {
		t.Errorf("Expected frame metadata to be `% x` but found `% x`", metadata, f.Metadata())
	}
	if len(f.Buf) != 22 {
		t.Errorf("Expected frame length to be %d but found %d", 22, len(f.Buf))
	}
}
