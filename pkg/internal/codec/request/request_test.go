package request_test

//import (
//	"testing"
//	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
//	"bytes"
//	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
//)
//
//func TestEncodeMetadata(t *testing.T) {
//	f := frame.Request(1, 0, header.FTRequestChannel, []byte{0x61}, []byte{0x61})
//
//	metadata := f.Metadata()
//
//	if !bytes.Equal(metadata, []byte{0x61}) {
//			t.Errorf("Expected metadata to be 0x61, found % x", metadata)
//	}
//}
