package proto

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/errorc"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"io"
	"sync"
	"sync/atomic"
)

// Note, there are two constituents to consider in dealing with concurrency here,
// one is the Transport side, which calls methods on Protocol.Handler.
// The other is the Application side. The Application side is sneaky, it trickles into
// many places; try and note entry point methods for Application goroutines.

type Protocol struct {
	// This is the application-provided description of behavior;
	// eg. we handle the plumbing in and out of this, this Handler
	// decides what messages to send and how to handle received ones.
	Handler *rs.RequestHandler

	out *output

	localSubscribers   map[uint32]rs.Subscriber
	localSubscriptions map[uint32]rs.Subscription
}

func NewProtocol(h *rs.RequestHandler, send func(*frame.Frame) error) *Protocol {
	if h == nil {
		panic("Cannot create protocol instance with a nil RequestHandler, please provice a non-nil handler.")
	}
	return &Protocol{
		Handler:            h,
		out:                &output{send: send, f: &frame.Frame{}},
		localSubscribers:   make(map[uint32]rs.Subscriber),
		localSubscriptions: make(map[uint32]rs.Subscription),
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
	case header.FTRequestResponse:
		p.handleRequestResponse(f)
	case header.FTRequestSubscription:
		p.handleRequestStream(f, p.Handler.HandleRequestSubscription(f))
	case header.FTRequestStream:
		p.handleRequestStream(f, p.Handler.HandleRequestStream(f))
	default:
		panic(fmt.Sprintf("Unknown frame: %s", f.Describe()))
	}
}
func (p *Protocol) Terminate() {
	for streamId, s := range p.localSubscribers {
		s.OnError(io.EOF)
		delete(p.localSubscribers, streamId)
	}
	for streamId, s := range p.localSubscriptions {
		s.Cancel()
		delete(p.localSubscriptions, streamId)
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
	var s = p.localSubscribers[f.StreamID()]
	if s == nil {
		// TODO: need to sort out protocol deal here
		return
	}
	s.OnNext(f)
}
func (p *Protocol) handleRequestN(f *frame.Frame) {
	var s = p.localSubscriptions[f.StreamID()]
	if s == nil {
		// TODO: need to sort out protocol deal here
		return
	}
	s.Request(int(requestn.RequestN(f)))
}
func (p *Protocol) handleRequestResponse(f *frame.Frame) {
	streamId := f.StreamID()
	out := p.Handler.HandleRequestResponse(f)
	out.Subscribe(&remoteRequestResponseSubscriber{
		streamId: streamId,
		out:      p.out,
	})
}
func (p *Protocol) handleRequestStream(f *frame.Frame, pub rs.Publisher) {
	var streamId = f.StreamID()
	var s = p.localSubscriptions[f.StreamID()]
	if s == nil {
		pub.Subscribe(&remoteSubscriber{
			streamId:           streamId,
			localSubscriptions: p.localSubscriptions,
			out:                p.out,
		})
		s = p.localSubscriptions[f.StreamID()]

		if s == nil {
			panic("Programming error: Provided RequestHandler#HandleChannel(..) returned a Publisher " +
				"that did not call OnSubscribe when Subscribed to. This is not supported.")
		}

		if n := request.InitialRequestN(f); n > 0 {
			s.Request(int(n))
		}
		return
	} else {
		panic(fmt.Sprintf("Protocol violation: %d is already a stream in use.", streamId))
	}
}
func (p *Protocol) handleRequestChannel(f *frame.Frame) {
	var streamId = f.StreamID()
	var subscription = p.localSubscriptions[streamId]
	var subscriber = p.localSubscribers[streamId]
	if subscriber == nil && subscription == nil {
		p.handleRequestStream(f, p.Handler.HandleChannel(rs.NewPublisher(func(s rs.Subscriber) {
			p.localSubscribers[streamId] = s
			s.OnSubscribe(&subscriptionToRemoteStream{
				streamId:   streamId,
				initial:    rs.CopyPayload(f),
				subscriber: s,
				out:        p.out,
			})
		})))
		return
	}

	if request.IsCompleteChannel(f) {
		subscriber.OnComplete()
		delete(p.localSubscribers, streamId)
	} else {
		subscriber.OnNext(f)
	}
}

// This is the applications subscription to the remote stream
type subscriptionToRemoteStream struct {
	started    int32
	streamId   uint32
	initial    rs.Payload
	subscriber rs.Subscriber
	out        *output
}

// Called by Application
func (r *subscriptionToRemoteStream) Request(n int) {
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
func (r *subscriptionToRemoteStream) Cancel() {
	panic("Cancel not yet implemented")
}

// Represents the remote subscriber - sending messages to this will have them delivered over
// the transport.
type remoteSubscriber struct {
	streamId           uint32
	localSubscriptions map[uint32]rs.Subscription
	out                *output
}

func (s *remoteSubscriber) OnSubscribe(subscription rs.Subscription) {
	s.localSubscriptions[s.streamId] = subscription
}
func (s *remoteSubscriber) OnNext(val rs.Payload) {
	s.out.sendResponse(s.streamId, val)
}
func (s *remoteSubscriber) OnError(err error) {
	s.out.sendError(s.streamId, err)
	delete(s.localSubscriptions, s.streamId)
}
func (s *remoteSubscriber) OnComplete() {
	s.out.sendResponseComplete(s.streamId)
	delete(s.localSubscriptions, s.streamId)
}

// Represents a remote request/response subscriber, waiting for its single response.
type remoteRequestResponseSubscriber struct {
	streamId uint32
	out      *output
}

func (s *remoteRequestResponseSubscriber) OnSubscribe(subscription rs.Subscription) {
	subscription.Request(1)
}
func (s *remoteRequestResponseSubscriber) OnNext(val rs.Payload) {
	s.out.sendResponseCompleteWithPayload(s.streamId, val)
}
func (s *remoteRequestResponseSubscriber) OnError(err error) {
	s.out.sendError(s.streamId, err)
}
func (s *remoteRequestResponseSubscriber) OnComplete() {}

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
func (out *output) sendResponseCompleteWithPayload(streamId uint32, val rs.Payload) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeResponse(out.f, streamId, header.FlagResponseComplete, val.Metadata(), val.Data())); err != nil {
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
