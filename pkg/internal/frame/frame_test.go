package frame_test

import (
	"bytes"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"testing"
)

var payloads = [][]byte{
	{},
	{1},
	{1, 2, 3, 4, 5, 6, 7},
}

func TestDecodeFrame(t *testing.T) {
	for _, payload := range payloads {
		buffer := &bytes.Buffer{}
		encoder := frame.NewFrameEncoder(buffer)
		decoder := frame.NewFrameDecoder(buffer)
		writeFrame, readFrame := &frame.Frame{}, &frame.Frame{}
		frame.EncodeSetup(writeFrame, 0, 60, 60, "test/test+meta", "test/test+data", []byte{}, payload)

		encoder.Write(writeFrame)
		decoder.Read(readFrame)

		if !bytes.Equal(readFrame.Data(), payload) {
			t.Errorf("Expected decoded payload to be % x, found % x", payload, readFrame.Data())
		}

	}
}
