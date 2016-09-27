package tcp

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/setup"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"net"
)

func ListenAndServe(address string, setup rs.ConnectionSetupHandler) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := server{listener, setup}
	return server.serve()
}

type server struct {
	listener net.Listener
	setup    rs.ConnectionSetupHandler
}

func (s *server) serve() error {
	for {
		rwc, err := s.listener.Accept()
		if err != nil {
			// TODO: See stdlib http loop, it checks for temporary network errors here
			return err
		}

		// TODO: Proper resource handling - close these guys on server close
		c := &conn{rwc: rwc, setup: s.setup}
		go c.serve()
	}
}

type conn struct {
	rwc      net.Conn
	frame    frame.Frame
	setup    rs.ConnectionSetupHandler
	protocol *proto.Protocol
	enc      *frame.FrameEncoder
	dec      *frame.FrameDecoder
}

func (c *conn) serve() {
	// TODO This should wrap in buffered io
	c.dec = frame.NewFrameDecoder(c.rwc)
	c.enc = frame.NewFrameEncoder(c.rwc)

	f := &c.frame

	// Handle Setup
	if err := c.dec.Read(f); err != nil {
		c.fatalError(err)
	}
	if f.Type() != header.FTSetup {
		c.fatalError(fmt.Errorf("Expected first frame to be SETUP, got %d.", f.Type()))
	}
	if setup.Version(f) != 0 {
		c.fatalError(fmt.Errorf("Expected version to be 0, got %d", setup.Version(f)))
	}

	sp := &setupPayload{
		payload:  f,
		mime:     setup.DataMimeType(f),
		metaMime: setup.MetadataMimeType(f),
	}

	handler, err := c.setup(sp, nil)
	if err != nil {
		c.fatalError(err)
	}

	c.protocol = &proto.Protocol{
		Handler: handler,
		Send: func(f *frame.Frame) error {
			fmt.Printf("[Server] %s\n", f.Describe())
			return c.enc.Write(f)
		},
	}

	for {
		if err := c.dec.Read(f); err != nil {
			fmt.Println("Failed to read frame; also, programmer failed to write error handling")
			panic(err)
		}

		fmt.Printf("[Client] %s\n", f.Describe())

		c.protocol.HandleFrame(f)
	}
}

func (c *conn) fatalError(err error) {
	fmt.Println("Programmer failed to write error handling")
	panic(err)
}

type setupPayload struct {
	payload  rs.Payload
	metaMime string
	mime     string
}

func (s *setupPayload) Data() []byte {
	return s.payload.Data()
}
func (s *setupPayload) Metadata() []byte {
	return s.payload.Metadata()
}
func (s *setupPayload) DataMimeType() string {
	return s.mime
}
func (s *setupPayload) MetadataMimeType() string {
	return s.metaMime
}
