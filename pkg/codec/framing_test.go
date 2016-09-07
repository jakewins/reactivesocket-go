package codec_test

import (
	"testing"
	"github.com/jakewins/reactivesocket-go/pkg/codec"
	"io"
)

var payloads = [][]byte{
  {},
	{1},
	{1,2,3,4,5,6,7},
}

func TestDecodeFrame(t *testing.T) {
	for _, test := range payloads  {

		piperead, pipewrite := io.Pipe()
		encoder := codec.FrameEncoder{pipewrite}
		decoder := codec.FrameDecoder{piperead}

		encoder.Write(test)
		decoder.Read()
	}
}