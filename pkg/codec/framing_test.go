package codec_test

import (
	"testing"
	"github.com/jakewins/reactivesocket-go/pkg/codec"
	"io"
	"bytes"
)

var payloads = [][]byte{
  {},
	{1},
	{1,2,3,4,5,6,7},
}

func TestDecodeFrame(t *testing.T) {
	for _, payload := range payloads  {
		buffer := &bytes.Buffer{}
		encoder := codec.FrameEncoder{buffer}
		decoder := codec.FrameDecoder{buffer}
		frame := &codec.Frame{}
		codec.SetupFrame(frame, 0, 60, 60, "test/test+meta", "test/test+data", []byte{}, payload)

		encoder.Write(frame)
		buffer.Next()
		decoder.Read(frame)

		if !bytes.Equal(frame.Data(), payload) {
			t.Errorf("Expected decoded payload to be % x, found % x", payload, frame.Data())
		}
	}
}