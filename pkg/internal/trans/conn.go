package trans

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/setup"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"io"
	"net"
)

// Exposes proto.Protocol over a net.Conn
type ReactiveConn struct {
	Id       int
	Rwc      net.Conn
	Setup    func(*ReactiveConn) (*rs.RequestHandler, error)
	frame    frame.Frame
	Protocol *proto.Protocol
	enc      *frame.FrameEncoder
	dec      *frame.FrameDecoder
}

// firstStreamId is used to start the stream id generator - you should set this
// to 2 if you are implementing a server and 1 if you are implementing a client,
// this maintains the odd/even invariant to separate clients and servers.
func (c *ReactiveConn) Initialize(firstStreamId uint32) {
	// TODO This should wrap in buffered io
	c.dec = frame.NewFrameDecoder(c.Rwc)
	c.enc = frame.NewFrameEncoder(c.Rwc)

	// Handle Setup
	handler, err := c.Setup(c)
	if err != nil {
		// TODO: Note that c.Setup can fail either with an application code error,
		// or with a network error from eg. ReadSetupFrame, below; we need to implement
		// some sort of scheme to track the state of the conn, handling errors based on
		// where we are at.
		c.fatalError(err)
	}

	c.Protocol = proto.NewProtocol(
		handler,
		firstStreamId,
		func(f *frame.Frame) error {
			fmt.Printf("[C%d] -> %s\n", c.Id, f.Describe())
			return c.enc.Write(f)
		},
	)
}

func (c *ReactiveConn) Serve() {
	f := &c.frame
	for {
		if err := c.dec.Read(f); err != nil {
			if err == io.EOF {
				fmt.Printf("[C%d] <- EOF\n", c.Id)
			}
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Err.Error() == "connection reset by peer" {
					fmt.Printf("[C%d] <- Connection Reset\n", c.Id)
				}
			}

			c.Protocol.Terminate(err)
		}

		fmt.Printf("[C%d] <- %s\n", c.Id, f.Describe())

		c.Protocol.HandleFrame(f)
	}
}

// When implementing a server, this reads the initial setup frame
func (c *ReactiveConn) ReadSetupFrame() (rs.ConnectionSetupPayload, error) {
	f := &c.frame
	if err := c.dec.Read(f); err != nil {
		return nil, err
	}
	if f.Type() != header.FTSetup {
		return nil, fmt.Errorf("Expected first frame to be SETUP, got %d.", f.Type())
	}
	if setup.Version(f) != 0 {
		return nil, fmt.Errorf("Expected version to be 0, got %d", setup.Version(f))
	}

	return rs.NewSetupPayload(
		setup.MetadataMimeType(f),
		setup.DataMimeType(f),
		f.Metadata(),
		f.Data(),
	), nil
}

// When implementing a client, this writes the initial setup frame
func (c *ReactiveConn) WriteSetupFrame(keepaliveInterval, maxLifetime uint32, setupPayload rs.ConnectionSetupPayload) error {
	f := &c.frame
	if err := c.enc.Write(frame.EncodeSetup(f, 0, keepaliveInterval, maxLifetime,
		setupPayload.MetadataMimeType(), setupPayload.DataMimeType(),
		setupPayload.Metadata(), setupPayload.Data())); err != nil {
		return err
	}
	return nil
}

func (c *ReactiveConn) fatalError(err error) {
	fmt.Println("Programmer failed to write error handling")
	c.Rwc.Close()
	panic(err) // TODO
}
