package tcp

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/trans"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"net"
)

// Connect to a TCP Reactive Socket identified by address
func Dial(address string, setup rs.ConnectionSetupPayload) (rs.ReactiveSocket, error) {
	return DialAndHandle(address, setup, &rs.RequestHandler{})
}

// Same as Dial, adding the ability to handle requests coming from the server
func DialAndHandle(address string, setup rs.ConnectionSetupPayload, handler *rs.RequestHandler) (rs.ReactiveSocket, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	rwc, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	c := &trans.ReactiveConn{
		Id:  0,
		Rwc: rwc,
		Setup: func(c *trans.ReactiveConn) (*rs.RequestHandler, error) {
			if err := c.WriteSetupFrame(1000, 0, setup); err != nil {
				return nil, err
			}

			return handler, nil
		},
	}

	c.Initialize(1)
	go c.Serve()

	return c.Protocol, nil
}
