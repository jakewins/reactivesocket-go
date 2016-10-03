package proto

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/errorc"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"sync"
	"sync/atomic"
)

// Note, there are two constituents to consider in dealing with concurrency here,
// one is the Transport side, which calls methods on Protocol.Handler.
// The other is the Application side. The Application side is sneaky, it trickles into
// many places; try and note entry point methods for Application goroutines.

type stream struct {
	// This is our sides subscribing to the remote publisher
	in rs.Subscriber

	// This is our sides subscription control for the remote subscriber
	out rs.Subscription
}

type Protocol struct {
	// This is the application-provided description of behavior;
	// eg. we handle the plumbing in and out of this, this Handler
	// decides what messages to send and how to handle received ones.
	Handler *rs.RequestHandler

	out *output

	// Only manipulated from HandleFrame, so no synchronization
	streams map[uint32]*stream
}

func NewProtocol(h *rs.RequestHandler, send func(*frame.Frame) error) *Protocol {
	if h == nil {
		panic("Cannot create protocol instance with a nil RequestHandler, please provice a non-nil handler.")
	}
	return &Protocol{
		Handler: h,
		streams: make(map[uint32]*stream),
		out:     &output{send: send, f: &frame.Frame{}},
	}
}

// This method is not goroutine safe
func (p *Protocol) HandleFrame(f *frame.Frame) {
	switch f.Type() {
	case header.FTRequestChannel:
		p.handleRequestChannel(f)
	case header.FTKeepAlive:
		p.handleKeepAlive(f)
	case header.FTResponse:
		p.handleResponse(f)
	case header.FTRequestN:
		p.handleRequestN(f)
	case header.FTFireAndForget:
		p.handleFireAndForget(f)
	default:
		panic(fmt.Sprintf("Unknown frame: %s", f.Describe()))
	}
}
func (p *Protocol) HandleEOF() {
	// Close all open streams
}
func (p *Protocol) handleRequestChannel(f *frame.Frame) {
	var streamId = f.StreamID()
	var theStream *stream = p.streams[streamId]
	if theStream == nil {
		theStream = p.createStream(streamId, f)
		if n := request.InitialRequestN(f); n > 0 {
			theStream.out.Request(int(n))
		}
		return
	}

	if request.IsCompleteChannel(f) {
		theStream.in.OnComplete()
	} else {
		theStream.in.OnNext(f)
	}
}
func (p *Protocol) handleFireAndForget(f *frame.Frame) {
	p.Handler.HandleFireAndForget(f)
}
func (p *Protocol) handleKeepAlive(f *frame.Frame) {
	if f.Flags()&header.FlagKeepaliveRespond != 0 {
		p.out.sendKeepAlive()
	}
}
func (p *Protocol) handleResponse(f *frame.Frame) {
	var stream = p.streams[f.StreamID()]
	if stream == nil {
		// TODO: need to sort out protocol deal here
		return
	}
	stream.in.OnNext(f)
}
func (p *Protocol) handleRequestN(f *frame.Frame) {
	var stream = p.streams[f.StreamID()]
	if stream == nil {
		// TODO: need to sort out protocol deal here
		return
	}
	stream.out.Request(int(requestn.RequestN(f)))
}

// TODO This whole *stream should be pooled on the Protocol instance and reused
func (p *Protocol) createStream(streamId uint32, initial *frame.Frame) *stream {
	newStream := &stream{}
	p.streams[streamId] = newStream

	var out = p.Handler.HandleChannel(rs.NewPublisher(func(s rs.Subscriber) {
		newStream.in = s
		s.OnSubscribe(&remoteStreamSubscription{
			streamId:   streamId,
			initial:    rs.CopyPayload(initial),
			subscriber: s,
			out:        p.out,
		})
	}))

	out.Subscribe(rs.NewSubscriber(
		func(subscription rs.Subscription) {
			newStream.out = subscription
		},
		func(val rs.Payload) {
			p.out.sendResponse(streamId, val)
		},
		func(err error) {
			p.out.sendError(streamId, err)
		},
		func() {
			// TODO: When do we clean up the stream reference?
			p.out.sendResponseComplete(streamId)
		},
	))

	if newStream.in == nil {
		panic("Programming error: Provided RequestHandler#HandleChannel(..) did not call " +
			"Subscribe on the provided publisher immediately when invoked. This is not supported.")
	}
	if newStream.out == nil {
		panic("Programming error: Provided RequestHandler#HandleChannel(..) returned a Publisher " +
			"that did not call OnSubscribe when Subscribed to. This is not supported.")
	}

	return newStream
}

// This is the applications subscription to the remote stream
type remoteStreamSubscription struct {
	started    int32
	streamId   uint32
	initial    rs.Payload
	subscriber rs.Subscriber
	out        *output
}

// Called by Application
func (r *remoteStreamSubscription) Request(n int) {
	// A bit precarious here; for efficiencies sake, the first payload
	// in a channel is bundled with the Request to start the channel.
	// Hence, the first req the App makes is immediately fulfilled.
	if atomic.CompareAndSwapInt32(&r.started, 0, 1) {
		r.subscriber.OnNext(r.initial)
		n -= 1
	}
	if n > 0 {
		r.out.sendRequestN(r.streamId, uint32(n))
	}
}

// Called by Application
func (r *remoteStreamSubscription) Cancel() {

}

// API to send outbound Frames. All methods on this struct can be expected to be called
// by both Application and Transport goroutines
type output struct {
	// Send any frame back to our remote counterpart.
	// Memory semantics here are that as soon as this method call
	// returns, the frame can be re-used. Any implementation of
	// Send must copy the contents of Frame if it wishes to retain
	// it beyond this call.
	send func(*frame.Frame) error

	// This is to allow concurrent Application threads per connection;
	// however, it'd be much better replaced by a ring buffer
	lock sync.Mutex
	f    *frame.Frame
}

func (out *output) sendResponse(streamId uint32, val rs.Payload) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeResponse(out.f, streamId, 0, val.Metadata(), val.Data())); err != nil {
		panic(err.Error()) // TODO
	}
}
func (out *output) sendError(streamId uint32, err error) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeError(out.f, streamId, errorc.ECApplicationError, nil, []byte(err.Error()))); err != nil {
		panic(err.Error()) // TODO
	}
}
func (out *output) sendResponseComplete(streamId uint32) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeResponse(out.f, streamId, header.FlagResponseComplete, nil, nil)); err != nil {
		panic(err.Error()) // TODO
	}
}
func (out *output) sendRequestN(streamId, n uint32) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeRequestN(out.f, streamId, n)); err != nil {
		panic(err.Error()) // TODO
	}
}
func (out *output) sendKeepAlive() {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeKeepalive(out.f, false)); err != nil {
		panic(err.Error()) // TODO
	}
}
