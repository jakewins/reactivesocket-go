package proto

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"fmt"
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
}

func (self *Protocol) HandleFrame(f *frame.Frame) {
	switch(f.Type()) {
	case header.FTRequestChannel:
		self.handleRequestChannel(f)
	case header.FTKeepAlive:
	// TODO
	default: panic(fmt.Sprintf("Unknown frame: %d", f.Type()))
	}
}
func (self *Protocol) handleRequestChannel(f *frame.Frame) {
	// uhhhh
	// so
	self.Handler.HandleChannel(f, nil)
}