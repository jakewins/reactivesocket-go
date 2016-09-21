package setup_test

import (
	"testing"
	"bytes"
	"github.com/jakewins/reactivesocket-go/pkg/frame"
	"github.com/jakewins/reactivesocket-go/pkg/frame/setup"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
)

func TestSetupFrameEncoding(t *testing.T) {
	var keepaliveInterval uint32 = 1
	var maxLifetime uint32 = 2
	var flags uint16 = setup.FlagStrictInterpretation
	metadataMime := "test/test+meta"
	metadata := []byte{1,2,3}
	data := []byte{4,5,6}
	dataMime := "test/test+data"

	f := &frame.Frame{}
	setup.Encode(f, flags, keepaliveInterval, maxLifetime, metadataMime, dataMime, metadata, data)

	if setup.Flags(f) & setup.FlagStrictInterpretation == 0 {
		t.Error("Expected frame to have metadata flag set")
	}
	if setup.Version(f) != 0 {
		t.Errorf("Expected version to be 0, found %d", setup.Version(f))
	}
	if setup.KeepaliveInterval(f) != keepaliveInterval {
		t.Errorf("Expected keepalive interval to be %d, found %d", keepaliveInterval, setup.KeepaliveInterval(f))
	}
	if setup.MaxLifetime(f) != maxLifetime {
		t.Errorf("Expected maxLifetime to be %d, found %d", maxLifetime, setup.MaxLifetime(f))
	}
	if setup.MetadataMimeType(f) != metadataMime {
		t.Errorf("Expected metadata mime type to be %s, found %s", metadataMime, setup.MetadataMimeType(f))
	}
	if setup.DataMimeType(f) != dataMime {
		t.Errorf("Expected data mime type to be %s, found %s", dataMime, setup.DataMimeType(f))
	}

	if f.Type() != header.FTSetup {
		t.Errorf("Expected type to be %d, found %d", header.FTSetup, f.Type())
	}
	if f.StreamID() != 0 {
		t.Errorf("Expected stream id to be 0, found %d", f.StreamID())
	}

	if !bytes.Equal(f.Data(), data) {
		t.Errorf("Expected frame data to be `% x` but found `% x`", data, f.Data())
	}
	if !bytes.Equal(f.Metadata(), metadata) {
		t.Errorf("Expected frame metadata to be `% x` but found `% x`", metadata, f.Metadata())
	}
}

func TestShouldFlagMetadataIfMetadataNonEmpty(t *testing.T) {
	t.Errorf("TODO")
}