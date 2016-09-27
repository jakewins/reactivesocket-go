package proto

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"sync"
)

type Protocol struct {
	// This is the application-provided description of behavior;
	// eg. we handle the plumbing in and out of this, this Handler
	// decides what messages to send and how to handle received ones.
	Handler *rs.RequestHandler

	// Send any frame back to our remote counterpart.
	// Memory semantics here are that as soon as this method call
	// returns, the frame can be re-used. Any implementation of
	// Send must copy the contents of Frame if it wishes to retain
	// it beyond this call.
	Send func(*frame.Frame)

	// Only manipulated from HandleFrame, so no synchronization
	streams map[uint32]rs.Subscriber

	// Protects f from concurrent application messages
	lock sync.Mutex

	// Used by when Send(..)-ing messages
	f *frame.Frame
}

func NewProtocol(h *rs.RequestHandler, send func(*frame.Frame)) *Protocol {
	if h == nil {
		panic("Cannot create protocol instance with a nil RequestHandler, please provice a non-nil handler.")
	}
	return &Protocol{
		Handler: h,
		Send:    send,
		streams: make(map[uint32]rs.Subscriber, 16),
		f:       &frame.Frame{},
	}
}

// This method is not goroutine safe
func (self *Protocol) HandleFrame(f *frame.Frame) {
	switch f.Type() {
	case header.FTRequestChannel:
		self.handleRequestChannel(f)
	case header.FTKeepAlive:
		self.handleKeepAlive(f)
	case header.FTResponse:
		self.handleResponse(f)
	default:
		panic(fmt.Sprintf("Unknown frame: %d", f.Type()))
	}
}
func (self *Protocol) handleRequestChannel(f *frame.Frame) {
	var streamId = f.StreamID()
	self.Handler.HandleChannel(f, rs.NewPublisher(func(s rs.Subscriber) {
		self.streams[streamId] = s
		s.OnSubscribe(rs.NewSubscription(func(n int) {
			self.lock.Lock()
			defer self.lock.Unlock()
			self.Send(frame.EncodeRequestN(self.f, streamId, uint32(n)))
		}, func() {

		}))
	}))
}
func (self *Protocol) handleKeepAlive(f *frame.Frame) {
	if f.Flags()&header.FlagKeepaliveRespond != 0 {
		self.Send(frame.EncodeKeepalive(self.f, false))
	}
}
func (self *Protocol) handleResponse(f *frame.Frame) {
	var subscriber = self.streams[f.StreamID()]
	if subscriber == nil {
		// TODO: need to sort out protocol deal here
		return
	}
	subscriber.OnNext(f)
}
