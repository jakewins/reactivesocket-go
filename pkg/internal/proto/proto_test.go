package proto_test

import (
	"bytes"
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/keepalive"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"testing"
)

type scenario struct {
	name      string
	handler   *rs.RequestHandler
	exchanges []exchange
}
type exchange struct {
	in  []*frame.Frame
	out []*frame.Frame
}

var scenarios = []scenario{
	scenario{"Simple keepalive(response plz) -> keepalive", nil, exchanges{
		exchange{
			in{keepalive.New(true)},
			out{keepalive.New(false)},
		},
	},
	},
	scenario{"Keepalive with no response", nil, exchanges{
		exchange{
			in{keepalive.New(false)},
			out{},
		},
	},
	},
}

type in []*frame.Frame
type out []*frame.Frame
type exchanges []exchange

func TestScenarios(t *testing.T) {
	for _, scenario := range scenarios {
		r := recorder{}
		p := proto.NewProtocol(nil, r.Record)

		for _, exchange := range scenario.exchanges {
			for _, f := range exchange.in {
				p.HandleFrame(f)
			}
			if err := r.AssertRecorded(exchange.out); err != nil {
				t.Errorf("Scenario `%s` failed: %s", scenario.name, err.Error())
			}
		}
	}
}

type recorder struct {
	recording []*frame.Frame
}

func (r *recorder) Record(f *frame.Frame) {
	r.recording = append(r.recording, f.Copy(nil))
}
func (r *recorder) Rewind() {
	r.recording = nil
}
func (r *recorder) AssertRecorded(expected []*frame.Frame) error {
	if len(expected) < len(r.recording) {
		for i := len(expected); i < len(r.recording); i++ {
			return fmt.Errorf("Expected no more than %d frames, frame %d is %s", len(expected), i+1, r.recording[i].Describe())
		}
		return nil
	}
	for idx, expect := range expected {
		if len(r.recording) <= idx {
			return fmt.Errorf("Expected frame %d to be %s, but there are no more recorded frames.", idx, expect.Describe())
		}
		found := r.recording[idx]
		if !bytes.Equal(found.Buf, expect.Buf) {
			return fmt.Errorf("Expected frame %d to be %s, found %s", idx, expect.Describe(), found.Describe())
		}
	}
	return nil
}
