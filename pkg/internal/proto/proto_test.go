package proto_test

import (
	"bytes"
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/keepalive"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/response"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"math"
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
type in []*frame.Frame
type out []*frame.Frame
type exchanges []exchange

var noopHandler = &rs.RequestHandler{}
var infinity = math.MaxInt64

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
		"RequestChannel where server requests some values, no ending", &rs.RequestHandler{
			HandleChannel: channelFactory(blackhole(2, infinity), sequencer(0, infinity)),
		}, exchanges{
			exchange{
				in{request.New(1337, 0, header.FTRequestChannel, nil, nil)},
				out{requestn.New(1337, 2)},
			},
			exchange{
				in{response.New(1337, 0, nil, nil)},
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
				t.Errorf("Scenario `%s` failed: \n%s", scenario.name, err.Error())
			}
			r.Rewind()
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
func sequencer(start, end int) func(rs.Payload) rs.Publisher {
	return func(init rs.Payload) rs.Publisher {
		var seq int = start
		return rs.NewPublisher(func(s rs.Subscriber) {
			s.OnSubscribe(rs.NewSubscription(
				func(n int) {
					if seq >= end {
						return
					}

					for ; seq < seq+n; seq++ {
						if seq >= end {
							s.OnComplete()
							return
						}
						s.OnNext(seq)
					}
				},
				func() {
					seq = end
					s.OnComplete()
				},
			))
		})
	}
}

// Creates publishers that request and discard in chunks specified by
// requestSize, cancelling the subscription if at cancelAt messages
func blackhole(requestSize, cancelAt int) func(rs.Publisher) {
	return func(source rs.Publisher) {
		var subscription rs.Subscription
		var remainingBeforeCancel = cancelAt
		var outstandingRequests = 0

		source.Subscribe(rs.NewSubscriber(
			func(s rs.Subscription) {
				s.Request(requestSize)
				outstandingRequests += requestSize
				subscription = s
			},
			func(v interface{}) {
				remainingBeforeCancel -= 1
				if remainingBeforeCancel <= 0 {
					subscription.Cancel()
					return
				}

				outstandingRequests -= 1
				if outstandingRequests <= 0 {
					outstandingRequests = min(requestSize, remainingBeforeCancel)
					subscription.Request(outstandingRequests)
				}
			}, nil, nil,
		))
	}
}
func channelFactory(in func(rs.Publisher), out func(rs.Payload) rs.Publisher) func(rs.Payload, rs.Publisher) rs.Publisher {
	return func(init rs.Payload, source rs.Publisher) rs.Publisher {
		in(source)
		return out(init)
	}
}

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}
