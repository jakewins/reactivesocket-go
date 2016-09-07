package setup_test

import (
	"testing"
	"github.com/jakewins/reactivesocket-go/pkg/codec/setup"
)

func TestSetupFrameEncoding(t *testing.T) {
	var keepaliveInterval uint32 = 1
	var maxLifetime uint32 = 2
	metadataMime := "test/test+meta"
	metadata := []byte{1,2,3}
	data := []byte{4,5,6}
	dataMime := "test/test+data"

	frame := setup.NewSetupFrame(0, keepaliveInterval, maxLifetime, metadataMime, dataMime, metadata, data)

	if setup.Flags(frame) & setup.FlagHasMetadata == 0 {
		t.Error("Expected frame to have metadata flag set")
	}
}

func TestShouldFlagMetadataIfMetadataNonEmpty(t *testing.T) {
	t.Errorf("TODO")
}