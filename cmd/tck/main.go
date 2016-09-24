package main

import (
	"flag"
	"os"
	"bufio"
	"log"
	"strings"
	"fmt"
	"net"
	"github.com/jakewins/reactivesocket-go/pkg/reactive"
	"github.com/jakewins/reactivesocket-go/pkg/frame"
	"strconv"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/frame/setup"
)

var (
	server bool
	host string
	port int
	file string
)

func init() {
	flag.BoolVar(&server, "server", false, "To launch the server")
	flag.StringVar(&host, "host", "localhost", "For the client only, determine host to connect to")
	flag.IntVar(&port, "port", 4567, "For client, port to connect to. For server, port to bind to")
	flag.StringVar(&file, "file", "", "Path to script file to run")
}

func main() {
	flag.Parse()

	if server {
		runServer(port, file)
	}
}

func runServer(port int, path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// This obv sucks. Just trying to wrap my head around how these files
	// work before refactoring. This is a map of data, metadata payload
	// it's used for matching against the first message received on a channel,
	// in order to decide which part of the TCK script to run
	channels := make(map[string]map[string][]string)
	requestResponseMarbles := make(map[string]map[string]string)
	requestStreamMarbles := make(map[string]map[string]string)
	requestSubscriptionMarbles := make(map[string]map[string]string)
	requestEchoChannel := make(map[string]map[string]bool)

	builder := NewRequestHandlerBuilder()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "%%")
		initData := parts[1]
		initMetadata := parts[2]
		fmt.Println(parts[0])
		switch parts[0] {
		case "rr":
			if requestResponseMarbles[initData] == nil {
				requestResponseMarbles[initData] = make(map[string]string)
			}
			requestResponseMarbles[initData][initMetadata] = parts[3]
		case "rs":
			if requestStreamMarbles[initData] == nil {
				requestStreamMarbles[initData] = make(map[string]string)
			}
			requestStreamMarbles[initData][initMetadata] = parts[3]
		case "sub":
			if requestSubscriptionMarbles[initData] == nil {
				requestSubscriptionMarbles[initData] = make(map[string]string)
			}
			requestSubscriptionMarbles[initData][initMetadata] = parts[3]
		case "echo":
			if requestEchoChannel[initData] == nil {
				requestEchoChannel[initData] = make(map[string]bool)
			}
			requestEchoChannel[initData][initMetadata] = true
		case "channel":
			if len(parts) == 5 {
				panic("Look into this, java handler had special case here: " + scanner.Text())
			}
			var script []string
			for scanner.Scan(); scanner.Text() != "}"; scanner.Scan() {
				script = append(script, scanner.Text())
			}
			if channels[initData] == nil {
				channels[initData] = make(map[string][]string)
			}
			channels[initData][initMetadata] = script
		default:
			panic("Unknown TCK server handler: " + parts[0])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	builder = builder.WithRequestResponse(requestResponseInitializer(requestResponseMarbles))

	log.Fatal(ListenAndServe(":" + strconv.Itoa(port), func(setup ConnectionSetupPayload, socket ReactiveSocket) (*RequestHandler, error) {
		return builder.Build(), nil
	}))
}

func requestResponseInitializer(stuff map[string]map[string]string) RequestResponseHandler {
	return func(firstPacket Payload) reactive.Publisher {
		return nil
	}
}


// Stuff that should be moved once hashed out

type Payload interface {
	Metadata() []byte
	Data() []byte
}

type ConnectionSetupPayload struct {
	metadata []byte
	data []byte
	metadataMime string
	dataMime string
}

func (self *ConnectionSetupPayload) Metadata() []byte {
	return self.metadata
}
func (self *ConnectionSetupPayload) Data() []byte {
	return self.data
}
func (self *ConnectionSetupPayload) MetadataMimeType() string {
	return self.metadataMime
}
func (self *ConnectionSetupPayload) DataMimeType() string {
	return self.dataMime
}

type RequestHandler struct {
	requestResponse RequestResponseHandler
}
func (h *RequestHandler) HandleRequestResponse(payload Payload) reactive.Publisher {
	return h.requestResponse(payload)
}

type RequestResponseHandler func(Payload) reactive.Publisher

func NewRequestHandlerBuilder() *RequestHandlerBuilder {
	return &RequestHandlerBuilder{}
}

type RequestHandlerBuilder struct {
	requestResponse RequestResponseHandler
}
func (s *RequestHandlerBuilder) WithRequestResponse(h RequestResponseHandler) *RequestHandlerBuilder {
	s.requestResponse = h
	return s
}

func (s *RequestHandlerBuilder) Build() *RequestHandler {
	return &RequestHandler{
		s.requestResponse,
	}
}

type ReactiveSocket interface {

}

// The ConnectionSetupHandler is invoked each time a new connection
// is established. It's purpose is to look at the setup payload the
// client sent, and return a RequestHandler of its choice to handle
// requests on the new connection. It is also given a ReactiveSocket
// instance, which can be used for server-initiated exchanges.
type ConnectionSetupHandler func(setup ConnectionSetupPayload, socket ReactiveSocket) (*RequestHandler, error)

// Transport

func ListenAndServe(address string, setup ConnectionSetupHandler) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := TCPReactiveSocketServer{listener, setup}
	return server.serve()
}

type Server interface {
	Addr() net.Addr
	Close() error
}

type TCPReactiveSocketServer struct {
	listener net.Listener
	setup ConnectionSetupHandler
}

func (s *TCPReactiveSocketServer) serve() error {
	for {
		fmt.Println("Waiting for connection..")
		rwc, err := s.listener.Accept()
		if err != nil {
			// TODO: See stdlib http loop, it checks for temporary network errors here
			return err
		}

		// TODO: Proper resource handling - close these guys on server close
		c := &conn{rwc:rwc, setup:s.setup}
		go c.serve()
	}
}

func (s *TCPReactiveSocketServer) Addr() net.Addr {
	return s.listener.Addr()
}

type conn struct {
	rwc net.Conn
	frame frame.Frame
	setup ConnectionSetupHandler
	enc *frame.FrameEncoder
	dec *frame.FrameDecoder
}

func (c *conn) serve() {
	c.dec = frame.NewFrameDecoder(c.rwc)
	f := c.frame

	// Handle Setup
	if err := c.dec.Read(&f); err != nil {
		c.fatalError(err)
	}

	if f.Type() != header.FTSetup {
		c.fatalError(fmt.Errorf("Expected first frame to be SETUP, got %d.", f.Type()))
	}

	_, err := c.setup(ConnectionSetupPayload{
		data: f.Data(),
		metadata: f.Metadata(),
		metadataMime: setup.MetadataMimeType(&f),
		dataMime: setup.DataMimeType(&f),
	}, nil)
	if err != nil {
		c.fatalError(err)
	}

	for {
		if err := c.dec.Read(&f); err != nil {
			fmt.Println("Failed to read frame; also, programmer failed to write error handling")
			panic(err)
		}

		fmt.Printf("%+v\n", f)
		fmt.Printf("Type: %d\n", f.Type())
	}
}

func (c *conn) fatalError(err error) {
	fmt.Println("Programmer failed to write error handling")
	panic(err)
}