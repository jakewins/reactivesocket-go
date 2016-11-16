package tcp

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"net"
)

func Dial(address string, setup rs.ConnectionSetupPayload) (rs.ReactiveSocket, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	rwc, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	c := &reactiveConn{
		id:  0,
		rwc: rwc,
		setup: func(c *reactiveConn) (*rs.RequestHandler, error) {
			f := &c.frame
			if err := c.enc.Write(frame.EncodeSetup(f, 0, 1000, 0,
				setup.MetadataMimeType(), setup.DataMimeType(),
				setup.Metadata(), setup.Data())); err != nil {
				c.fatalError(err)
			}

			return &rs.RequestHandler{}, nil
		},
	}

	c.initialize(1)
	go c.serve()

	return c.protocol, nil
}
