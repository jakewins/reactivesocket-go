package proto_test

import (
	"testing"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/keepalive"
	"bytes"
)

func TestKeepAliveRespond(t *testing.T) {
	r := recorder{}
	p := proto.NewProtocol(nil, r.Record)

	p.HandleFrame(keepalive.New(true))

	r.Expect(t, keepalive.New(false))
}


func TestKeepAliveDontRespond(t *testing.T) {
	r := recorder{}
	p := proto.NewProtocol(nil, r.Record)

	p.HandleFrame(keepalive.New(false))

	r.Expect(t) // Eg. no responses
}

type recorder struct {
	recording []*frame.Frame
}
func (r *recorder) Record(f *frame.Frame) {
	r.recording = append(r.recording, f.Copy(nil))
}
func (r *recorder) Expect(t *testing.T, expected ...*frame.Frame) {
	if len(expected) < len(r.recording) {
		for i:=len(expected); i<len(r.recording); i++ {
			t.Errorf("Expected no more than %d frames, frame %d is %s", len(expected), i+1, r.recording[i].Describe())
		}
		return
	}
	for idx, expect := range expected {
		if len(r.recording) <= idx {
			t.Errorf("Expected frame %d to be %s, but there are no more recorded frames.", idx, expect.Describe())
			continue
		}
		found := r.recording[idx]
		if !bytes.Equal(found.Buf, expect.Buf) {
			t.Errorf("Expected frame %d to be %s, found %s", idx, expect.Describe(), found.Describe())
		}
	}
}