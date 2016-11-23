package ws

import (
	"errors"
	"github.com/jakewins/reactivesocket-go/pkg/internal/trans"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"github.com/jakewins/reactivesocket-go/pkg/transport"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func Listen(address string, setup rs.ConnectionSetupHandler) (transport.Server, error) {
	laddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}

	return &wssServer{
		listener: &interruptibleListener{
			listener,
			make(chan int, 2),
		},
		setup:           setup,
		shutdownWaiters: &sync.WaitGroup{},
	}, nil
}

type wssServer struct {
	listener        *interruptibleListener
	setup           rs.ConnectionSetupHandler
	shutdownWaiters *sync.WaitGroup
}

func (s *wssServer) Serve() error {
	defer s.shutdownWaiters.Done()
	defer s.listener.Close()

	var connIds int64 = 0
	h := &websocket.Server{
		Handler: func(rwc *websocket.Conn) {
			connId := atomic.AddInt64(&connIds, 1) - 1
			c := &trans.ReactiveConn{
				Id:    int(connId),
				Rwc:   rwc,
				Setup: s.setupConnection,
			}
			go func() {
				c.Initialize(2)
				c.Serve()
			}()
		},
	}
	httpServer := &http.Server{
		Addr:    s.listener.Addr().String(),
		Handler: h,
	}

	err := httpServer.Serve(s.listener)
	if err == shutdownToken {
		return nil
	}
	return err
}
func (s *wssServer) Shutdown() {
	s.listener.shutdown()
}
func (s *wssServer) AwaitShutdown() {
	s.shutdownWaiters.Wait()
}
func (s *wssServer) setupConnection(c *trans.ReactiveConn) (*rs.RequestHandler, error) {
	sp, err := c.ReadSetupFrame()
	if err != nil {
		return nil, err
	}
	return s.setup(sp, c.Protocol)
}

var shutdownToken = errors.New("induced shutdown")

// HTTP server doesn't have a clean shutdown mechanism, so we inject errors
// into the accept loop to stop it.
type interruptibleListener struct {
	*net.TCPListener
	control chan int
}

func (l *interruptibleListener) Accept() (net.Conn, error) {
	for {
		l.SetDeadline(time.Now().Add(time.Second))

		newConn, err := l.TCPListener.Accept()

		select {
		case <-l.control:
			return nil, shutdownToken
		default:
		}

		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}

		return newConn, err
	}
}
func (l *interruptibleListener) shutdown() {
	close(l.control)
}
