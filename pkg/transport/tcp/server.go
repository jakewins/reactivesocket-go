package tcp

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/trans"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"net"
	"sync"
	"time"
)

func Listen(address string, setup rs.ConnectionSetupHandler) (Server, error) {
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
		control:         make(chan string, 2),
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
	defer s.listener.Close()

	var connIds int = 0
	for {
		if s.checkForShutdown() {
			return nil
		}
		s.listener.SetDeadline(time.Now().Add(1e9))
		rwc, err := s.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Temporary() {
				continue
			}
			return err
		}

		// TODO: Proper resource handling - close these guys on server close
		connIds += 1
		c := &trans.ReactiveConn{
			Id:    connIds,
			Rwc:   rwc,
			Setup: s.setupConnection,
		}
		go func() {
			c.Initialize(2)
			c.Serve()
		}()
	}
}
func (s *server) Shutdown() {
	close(s.control)
}
func (s *server) AwaitShutdown() {
	s.shutdownWaiters.Wait()
}
func (s *server) setupConnection(c *trans.ReactiveConn) (*rs.RequestHandler, error) {
	sp, err := c.ReadSetupFrame()
	if err != nil {
		return nil, err
	}
	return s.setup(sp, c.Protocol)
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
