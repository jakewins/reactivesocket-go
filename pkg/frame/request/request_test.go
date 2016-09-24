package request_test

import (
	"testing"
	"bytes"
	"github.com/jakewins/reactivesocket-go/pkg/frame"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/frame/request"
)

func TestFireAndForgetFrameEncoding(t *testing.T) {
	var flags uint16 = 0
	metadata := []byte{1,2,3}
	data := []byte{4,5,6}
	var streamId uint32 = 1337

	f := &frame.Frame{}
	request.Encode(f, streamId, flags, header.FTFireAndForget, metadata, data)

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
		t.Errorf("Expected frame length to be %d but found %d", 60, len(f.Buf))
	}
}

func TestRequestResponseFrameEncoding(t *testing.T) {
	var flags uint16 = 0
	metadata := []byte{1,2,3}
	data := []byte{4,5,6}
	var streamId uint32 = 1337

	f := &frame.Frame{}
	request.Encode(f, streamId, flags, header.FTRequestResponse, metadata, data)

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
		t.Errorf("Expected frame length to be %d but found %d", 60, len(f.Buf))
	}
}