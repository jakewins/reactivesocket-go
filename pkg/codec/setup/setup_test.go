package setup_test

import (
	"testing"
	"bytes"
	"github.com/jakewins/reactivesocket-go/pkg/codec"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/codec/setup"
)

func TestSetupFrameEncoding(t *testing.T) {
	var keepaliveInterval uint32 = 1
	var maxLifetime uint32 = 2
	var flags uint16 = setup.SetupFlagStrictInterpretation
	metadataMime := "test/test+meta"
	metadata := []byte{1,2,3}
	data := []byte{4,5,6}
	dataMime := "test/test+data"

	frame := &codec.Frame{}
	codec.SetupFrame(frame, flags, keepaliveInterval, maxLifetime, metadataMime, dataMime, metadata, data)
	buf := frame.Buf

	if setup.Flags(buf) & setup.SetupFlagStrictInterpretation == 0 {
		t.Error("Expected frame to have metadata flag set")
	}
	if setup.Version(buf) != 0 {
		t.Errorf("Expected version to be 0, found %d", setup.Version(buf))
	}
	if setup.KeepaliveInterval(buf) != keepaliveInterval {
		t.Errorf("Expected keepalive interval to be %d, found %d", keepaliveInterval, setup.KeepaliveInterval(buf))
	}
	if setup.MaxLifetime(buf) != maxLifetime {
		t.Errorf("Expected maxLifetime to be %d, found %d", maxLifetime, setup.MaxLifetime(buf))
	}
	if setup.MetadataMimeType(buf) != metadataMime {
		t.Errorf("Expected metadata mime type to be %s, found %s", metadataMime, setup.MetadataMimeType(buf))
	}
	if setup.DataMimeType(buf) != dataMime {
		t.Errorf("Expected data mime type to be %s, found %s", dataMime, setup.DataMimeType(buf))
	}

	if frame.Type() != header.FTSetup {
		t.Errorf("Expected type to be %d, found %d", header.FTSetup, frame.Type())
	}
	if frame.StreamID() != 0 {
		t.Errorf("Expected stream id to be 0, found %d", frame.StreamID())
	}

	if !bytes.Equal(frame.Data(), data) {
		t.Errorf("Expected frame data to be `% x` but found `% x`", data, frame.Data())
	}
	if !bytes.Equal(frame.Metadata(), metadata) {
		t.Errorf("Expected frame metadata to be `% x` but found `% x`", metadata, frame.Metadata())
	}
}

func TestShouldFlagMetadataIfMetadataNonEmpty(t *testing.T) {
	t.Errorf("TODO")
}