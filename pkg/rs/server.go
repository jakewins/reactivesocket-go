package rs

// The ConnectionSetupHandler is invoked each time a new connection
// is established. It's purpose is to look at the setup payload the
// client sent, and return a RequestHandler of its choice to handle
// requests on the new connection. It is also given a ReactiveSocket
// instance, which can be used for server-initiated exchanges.
type ConnectionSetupHandler func(setup ConnectionSetupPayload, socket ReactiveSocket) (*RequestHandler, error)

type ConnectionSetupPayload interface {
	Payload
	MetadataMimeType() string
	DataMimeType() string
}

type RequestHandler struct {
	HandleRequestResponse     func(Payload) Publisher
	HandleRequestStream       func(Payload) Publisher
	HandleRequestSubscription func(Payload) Publisher
	HandleChannel             func(Publisher) Publisher
	HandleFireAndForget       func(Payload)
	HandleMetadataPush        func(Payload)
}
