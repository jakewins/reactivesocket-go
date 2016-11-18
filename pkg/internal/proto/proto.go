package proto

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/errorc"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/requestn"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/response"
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

	nextStreamId uint32
}

func NewProtocol(h *rs.RequestHandler, firstStreamId uint32, send func(*frame.Frame) error) *Protocol {
	if h == nil {
		panic("Cannot create protocol instance with a nil RequestHandler, please provice a non-nil handler.")
	}
	return &Protocol{
		Handler:            h,
		out:                &output{send: send, f: &frame.Frame{}},
		localSubscribers:   make(map[uint32]rs.Subscriber),
		localSubscriptions: make(map[uint32]rs.Subscription),
		nextStreamId:       firstStreamId,
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
	case header.FTMetadataPush:
		p.handleMetadataPush(f)
	case header.FTError:
		p.handleError(f)
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

func (p *Protocol) FireAndForget(initial rs.Payload) rs.Publisher {
	streamId := p.generateStreamId()
	p.out.sendRequest(streamId, header.FTFireAndForget, initial)
	return rs.NewEmptyPublisher()
}
func (p *Protocol) RequestStream(initial rs.Payload) rs.Publisher {
	streamId := p.generateStreamId()
	initial = rs.CopyPayload(initial)
	return p.createPublisherForRemoteStream(streamId, func(n int, sub rs.Subscriber) int {
		p.out.sendRequestWithInitialN(streamId, uint32(n), header.FTRequestStream, initial)
		return 0
	})
}
func (p *Protocol) RequestSubscription(initial rs.Payload) rs.Publisher {
	streamId := p.generateStreamId()
	initial = rs.CopyPayload(initial)
	return p.createPublisherForRemoteStream(streamId, func(n int, sub rs.Subscriber) int {
		p.out.sendRequestWithInitialN(streamId, uint32(n), header.FTRequestSubscription, initial)
		return 0
	})
}
func (p *Protocol) RequestResponse(initial rs.Payload) rs.Publisher {
	streamId := p.generateStreamId()
	initial = rs.CopyPayload(initial)
	return p.createPublisherForRemoteStream(streamId, func(n int, sub rs.Subscriber) int {
		p.out.sendRequest(streamId, header.FTRequestResponse, initial)
		return 0
	})
}
func (p *Protocol) RequestChannel(payloads rs.Publisher) rs.Publisher {
	streamId := p.generateStreamId()
	return p.createPublisherForRemoteStream(streamId, func(n int, sub rs.Subscriber) int {
		payloads.Subscribe(&requesterRemoteSubscriber{
			streamId:           streamId,
			localSubscriptions: p.localSubscriptions,
			out:                p.out,
			initialRequestN:    uint32(n),
			isFirstPayload:     true,
		})
		return 0
	})
}
func (p *Protocol) generateStreamId() uint32 {
	for {
		candidate := p.nextStreamId
		if atomic.CompareAndSwapUint32(&p.nextStreamId, candidate, candidate+2) {
			return candidate
		}
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
	if response.IsCompleteStream(f) {
		s.OnComplete()
		delete(p.localSubscribers, f.StreamID())
	} else {
		s.OnNext(f)
	}
}
func (p *Protocol) handleRequestN(f *frame.Frame) {
	var s = p.localSubscriptions[f.StreamID()]
	if s == nil {
		// TODO: need to sort out protocol deal here
		return
	}
	s.Request(int(requestn.RequestN(f)))
}
func (p *Protocol) handleError(f *frame.Frame) {
	var s = p.localSubscribers[f.StreamID()]
	if s == nil {
		// TODO: need to sort out protocol deal here
		fmt.Printf("%s", f.Describe())
		return
	}
	s.OnError(fmt.Errorf("Error %d: %s", errorc.ErrorCode(f.Buf), string(f.Data())))
}
func (p *Protocol) handleMetadataPush(f *frame.Frame) {
	p.Handler.HandleMetadataPush(f)
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
		pub.Subscribe(&responderRemoteSubscriber{
			streamId:           streamId,
			localSubscriptions: p.localSubscriptions,
			out:                p.out,
		})
		s = p.localSubscriptions[f.StreamID()]

		if s == nil {
			panic("Programming error: Provided RequestHandler#HandleXXX(..) returned a Publisher " +
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
		firstMessage := rs.CopyPayload(f)
		p.handleRequestStream(f, p.Handler.HandleChannel(
			p.createPublisherForRemoteStream(streamId, func(n int, sub rs.Subscriber) int {
				sub.OnNext(firstMessage)
				return n - 1
			})))
		return
	}

	if request.IsCompleteStream(f) {
		subscriber.OnComplete()
		delete(p.localSubscribers, streamId)
	} else {
		subscriber.OnNext(f)
	}
}
func (p *Protocol) createPublisherForRemoteStream(streamId uint32, onFirstRequestN func(int, rs.Subscriber) int) rs.Publisher {
	return rs.NewPublisher(func(s rs.Subscriber) {
		p.localSubscribers[streamId] = s
		s.OnSubscribe(&subscriptionToRemoteStream{
			streamId:        streamId,
			onFirstRequestN: onFirstRequestN,
			subscriber:      s,
			out:             p.out,
		})
	})
}

// This is the applications subscription to the remote stream
type subscriptionToRemoteStream struct {
	streamId        uint32
	onFirstRequestN func(int, rs.Subscriber) int
	subscriber      rs.Subscriber
	out             *output
}

// Called by Application
func (r *subscriptionToRemoteStream) Request(n int) {
	// A bit precarious here; for efficiencies sake, the first payload
	// in a channel is bundled with the Request to start the channel.
	// Hence, the first req the App makes is immediately fulfilled.
	if r.onFirstRequestN != nil {
		onFirstRequest := r.onFirstRequestN
		r.onFirstRequestN = nil
		n = onFirstRequest(n, r.subscriber)
	}
	if n > 0 {
		r.out.sendRequestN(r.streamId, uint32(n))
	}
}

// Called by Application
func (r *subscriptionToRemoteStream) Cancel() {
	r.out.sendCancel(r.streamId)
}

// Represents the remote subscriber - sending messages to this will have them delivered over
// the transport.
type responderRemoteSubscriber struct {
	streamId           uint32
	localSubscriptions map[uint32]rs.Subscription
	out                *output
}

func (s *responderRemoteSubscriber) OnSubscribe(subscription rs.Subscription) {
	s.localSubscriptions[s.streamId] = subscription
}
func (s *responderRemoteSubscriber) OnNext(val rs.Payload) {
	s.out.sendResponse(s.streamId, val)
}
func (s *responderRemoteSubscriber) OnError(err error) {
	s.out.sendError(s.streamId, err)
	delete(s.localSubscriptions, s.streamId)
}
func (s *responderRemoteSubscriber) OnComplete() {
	s.out.sendResponseComplete(s.streamId)
	delete(s.localSubscriptions, s.streamId)
}

type requesterRemoteSubscriber struct {
	streamId           uint32
	localSubscriptions map[uint32]rs.Subscription
	out                *output
	initialRequestN    uint32
	isFirstPayload     bool
}

func (s *requesterRemoteSubscriber) OnSubscribe(subscription rs.Subscription) {
	s.localSubscriptions[s.streamId] = subscription
	subscription.Request(1)
}
func (s *requesterRemoteSubscriber) OnNext(val rs.Payload) {
	if s.isFirstPayload {
		s.isFirstPayload = false
		s.out.sendRequestWithInitialN(s.streamId, s.initialRequestN, header.FTRequestChannel, val)
	} else {
		s.out.sendRequest(s.streamId, header.FTRequestChannel, val)
	}
}
func (s *requesterRemoteSubscriber) OnError(err error) {
	s.out.sendError(s.streamId, err)
	delete(s.localSubscriptions, s.streamId)
}
func (s *requesterRemoteSubscriber) OnComplete() {
	s.out.sendRequestComplete(s.streamId)
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
func (out *output) sendRequest(streamId uint32, frameType uint16, val rs.Payload) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeRequest(out.f, streamId, 0, frameType, val.Metadata(), val.Data())); err != nil {
		panic(err.Error()) // TODO
	}
}
func (out *output) sendRequestComplete(streamId uint32) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeRequest(out.f, streamId, header.FlagRequestChannelComplete,
		header.FTRequestChannel, nil, nil)); err != nil {
		panic(err.Error()) // TODO
	}
}
func (out *output) sendRequestWithInitialN(streamId, initialN uint32, frameType uint16, val rs.Payload) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeRequestWithInitialN(out.f, streamId, initialN, 0, frameType, val.Metadata(), val.Data())); err != nil {
		panic(err.Error()) // TODO
	}
}
func (out *output) sendCancel(streamId uint32) {
	out.lock.Lock()
	defer out.lock.Unlock()
	if err := out.send(frame.EncodeCancel(out.f, streamId)); err != nil {
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
