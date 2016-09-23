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

	builder := NewRequestHandlerBuilder()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "%%")
		fmt.Println(parts[0])
		switch parts[0] {
		case "rr":
			builder = builder.WithRequestResponse(requestResponseInitializer(parts[1], parts[2], parts[3]))
		default:
			panic("Unknown TCK server handler: " + parts[0])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Fatal(ListenAndServe(":" + strconv.Itoa(port), builder.Build()))
}

func requestResponseInitializer(initialData, initialMetaData, marble string) RequestResponseHandler {
	return func(firstPacket Payload) reactive.Publisher {
		return nil
	}
}


// Stuff that should be moved once hashed out

type Payload interface {
	Metadata() []byte
	Data() []byte
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

func (s *RequestHandlerBuilder) Build() RequestHandler {
	return RequestHandler{
		s.requestResponse,
	}
}

// Transport

func ListenAndServe(address string, handler RequestHandler) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := TCPReactiveSocketServer{listener}
	return server.serve()
}

type Server interface {
	Addr() net.Addr
	Close() error
}

type TCPReactiveSocketServer struct {
	listener net.Listener
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
		c := &conn{rwc:rwc}
		go c.serve()
	}
}

func (s *TCPReactiveSocketServer) workerLoop(conn net.Conn) {

}

func (s *TCPReactiveSocketServer) Addr() net.Addr {
	return s.listener.Addr()
}


type conn struct {
	rwc net.Conn
	frame *frame.Frame
	enc *frame.FrameEncoder
	dec *frame.FrameDecoder
}

func (c *conn) serve() {
	c.dec = frame.NewFrameDecoder(c.rwc)

	for {
		if err := c.dec.Read(c.frame); err != nil {
			fmt.Println("Failed to read frame; also, programmer failed to write error handling")
			panic(err)
		}

		fmt.Printf("%+v", c.frame)
	}
}