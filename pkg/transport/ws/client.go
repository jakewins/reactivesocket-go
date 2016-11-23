package ws

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/internal/trans"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"golang.org/x/net/websocket"
)

// Connect to a TCP Reactive Socket identified by address, formatted as hostname:port
func Dial(address string, setup rs.ConnectionSetupPayload) (rs.ReactiveSocket, error) {
	return DialAndHandle(address, setup, &rs.RequestHandler{})
}

// Same as Dial, adding the ability to handle requests coming from the server
func DialAndHandle(address string, setup rs.ConnectionSetupPayload, handler *rs.RequestHandler) (rs.ReactiveSocket, error) {
	rwc, err := websocket.Dial(fmt.Sprintf("ws://%s/ws", address), "", fmt.Sprintf("http://%s/", address))
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
