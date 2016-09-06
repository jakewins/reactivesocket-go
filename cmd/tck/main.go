package main

import (
	"flag"
	"os"
	"bufio"
	"log"
	"strings"
	"fmt"
	"io"
	"net"
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

	builder := NewServerBuilder()

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


	server, err := ServeReactiveSocketOverTCP(":" + port, builder.Build())
	defer server.Close()
}

func requestResponseInitializer(initialData, initialMetaData, marble string) *func(Payload) Publisher {
	return func(firstPacket Payload) Publisher {
		return nil
	}
}


// Stuff that should be moved once hashed out

type Subscriber interface {

}

type Publisher interface {
	Subscribe(s Subscriber)
}

type RequestHandler interface {
	HandleRequestResponse(initial Payload) Publisher
}

type Initializer func(Payload) Publisher

type Payload interface {
	Metadata() io.Reader
	Data() io.Reader
}

func NewServerBuilder() ServerBuilder {
	return ServerBuilder{}
}

type ServerBuilder struct {

}
func (self *ServerBuilder) WithRequestResponse(initializer Initializer) ServerBuilder
func (self *ServerBuilder) Build() RequestHandler

// Transport

func ServeReactiveSocketOverTCP(address string, handler RequestHandler) (Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := TCPReactiveSocketServer{listener}
	go server.acceptLoop()

	return server, nil
}

type Server interface {
	Addr() net.Addr
	Close() error
}

type TCPReactiveSocketServer struct {
	listener net.Listener
}

func (s *TCPReactiveSocketServer) acceptLoop() {
	for {
		fmt.Println("Waiting for connection..")
		conn, err := s.listener.Accept()
		if err != nil {
			// TODO: Explore client-provided error strategies and best practices
			log.Println("Failure accepting new inbound request, server will shut down: %s", err.Error())
			return
		}

		fmt.Println("Got one, spinning off worker thread!")
		// TODO: Proper resource handling - close these guys on server close
		go s.workerLoop(conn)
	}
}

func (s *TCPReactiveSocketServer) workerLoop(conn net.Conn) {

}

func (s *TCPReactiveSocketServer) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *TCPReactiveSocketServer) Close() error {
	// TODO: This will trigger an error in the accept loop, rewrite this to use a flag or something
	return s.listener.Close()
}

// Protocol

type FrameLengthDecoder struct {
	source *io.Reader
}

// Options; pass a full buffer (or view..) upstream
//          let upstream provide a buffer to read from, optionally having to read more

// Noting that flatbuffers uses full buffers.. and actually, the subscriber model kinda mandates it as well
// So.. I guess FrameLengthDecoder should spit out whole bytes