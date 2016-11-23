package ws_test

import (
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"github.com/jakewins/reactivesocket-go/pkg/transport/ws"
)

func ExampleClient() {
	socket, err := ws.Dial("localhost:5678", rs.NewSetupPayload("text/json", "text/json", nil, nil))

	if err != nil {
		panic(err)
	}

	stream := socket.RequestSubscription(rs.NewPayload([]byte{1, 2, 3}, []byte{4, 5, 5}))
	stream.Subscribe(rs.NewSubscriber(func(s rs.Subscription) {
		s.Request(100)
	}, func(next rs.Payload) {
		// Handle payload
	}, func(err error) {
		// Handle errors..
	}, func() {
		// On complete
	}))
}

func ExampleServer() {
	server, err := ws.Listen(":0", func(setup rs.ConnectionSetupPayload, socket rs.ReactiveSocket) (*rs.RequestHandler, error) {
		// Choose a request handler based on the setup payload. Each connection has its own handler.
		return &rs.RequestHandler{
			// See rs.RequestHandler for many other request handler types
			HandleRequestSubscription: func(initial rs.Payload) rs.Publisher {
				// Return any stream you like, based on the initial payload
				return rs.NewEmptyPublisher()
			},
		}, nil
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting server!")
	go server.Serve()

	fmt.Println("Shutting down..")
	server.Shutdown()
	server.AwaitShutdown()
	// Output:
	// Starting server!
	// Shutting down..
}
