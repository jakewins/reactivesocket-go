package tcp

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame/setup"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"io"
	"net"
	"sync"
	"time"
)

func NewServer(address string, setup rs.ConnectionSetupHandler) (Server, error) {
	laddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}

	s := &server{
		listener:        listener,
		setup:           setup,
		control:         make(chan string, 8),
		shutdownWaiters: &sync.WaitGroup{},
	}
	s.shutdownWaiters.Add(1)
	return s, nil
}

type Server interface {
	// Runs the accept loop for this server, returns when the server is shut down.
	Serve() error
	// Signal the accept loop to shut down
	Shutdown()
	// Block until the server shuts down
	AwaitShutdown()
}

type server struct {
	listener        *net.TCPListener
	setup           rs.ConnectionSetupHandler
	control         chan string
	shutdownWaiters *sync.WaitGroup
}

func (s *server) Serve() error {
	defer s.shutdownWaiters.Done()
	var connIds int = 0
	for {
		if s.checkForShutdown() {
			return nil
		}
		s.listener.SetDeadline(time.Now().Add(1e9))
		rwc, err := s.listener.Accept()
		if err != nil {
			// TODO: See stdlib http loop, it checks for temporary network errors here
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			return err
		}

		// TODO: Proper resource handling - close these guys on server close
		connIds += 1
		c := &conn{
			id:    connIds,
			rwc:   rwc,
			setup: s.setup,
		}
		go c.serve()
	}
}
func (s *server) Shutdown() {
	close(s.control)
}
func (s *server) AwaitShutdown() {
	s.shutdownWaiters.Wait()
}
func (s *server) checkForShutdown() bool {
	select {
	case <-s.control:
		s.listener.Close()
		return true
	default:
		return false
	}
}

type conn struct {
	id       int
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

	c.protocol = proto.NewProtocol(
		handler,
		func(f *frame.Frame) error {
			fmt.Printf("[Server C%d] %s\n", c.id, f.Describe())
			return c.enc.Write(f)
		},
	)

	for {
		if err := c.dec.Read(f); err != nil {
			if err == io.EOF {
				fmt.Printf("[Client C%d] EOF\n", c.id)
				c.protocol.HandleEOF()
				return
			}
			fmt.Println("Failed to read frame; also, programmer failed to write error handling")
			panic(err) // TODO
		}

		fmt.Printf("[Client C%d] %s\n", c.id, f.Describe())

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
