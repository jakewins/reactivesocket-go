package proto_test

import (
	"bytes"
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/keepalive"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"testing"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
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
type in []*frame.Frame
type out []*frame.Frame
type exchanges []exchange

var noopHandler = &rs.RequestHandler{}

var scenarios = []scenario{
	{
		"Simple keepalive(response plz) -> keepalive", noopHandler, exchanges{
			exchange{
				in{keepalive.New(true)},
				out{keepalive.New(false)},
			},
		},
	},
	{
		"Keepalive with no response", noopHandler, exchanges{
			exchange{
				in{keepalive.New(false)},
				out{},
			},
		},
	},
	{
		"RequestChannel", &rs.RequestHandler{
		  HandleChannel: channelFactory(blackhole, sequencer),
		}, exchanges{
			exchange{
				in{request.New(1, 0, header.FTRequestChannel, nil, nil)},
				out{},
			},
		},
	},
}

func TestScenarios(t *testing.T) {
	for _, scenario := range scenarios {
		r := recorder{}
		p := proto.NewProtocol(scenario.handler, r.Record)

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

// Creates channels that spit out incrementing sequences of numbers
func sequencer(init rs.Payload) rs.Publisher {
	var seq int = 0
	return rs.NewPublisher(func(s rs.Subscriber) {
			s.OnSubscribe((&rs.SubscriptionParts{
				Request:func(n int) {
					for ; seq < seq + n; seq++ {
						s.OnNext(seq)
					}
				},
				Cancel:func() {
					s.OnComplete()
				},
			}).Build())
	})
}
// Creates publishers that indefinitely request and discards messages
func blackhole(source rs.Publisher) {
	var subscription rs.Subscription
	source.Subscribe((&rs.SubscriberParts{
		OnSubscribe: func(s rs.Subscription) {
			s.Request(1)
			subscription = s
		},
		OnNext: func(v interface{}) {
			subscription.Request(1)
		},
	}).Build())
}
func channelFactory(in func(rs.Publisher), out func(rs.Payload) rs.Publisher) func(rs.Payload, rs.Publisher) rs.Publisher {
	return func(init rs.Payload, source rs.Publisher) rs.Publisher {
		in(source)
		return out(init)
	}
}