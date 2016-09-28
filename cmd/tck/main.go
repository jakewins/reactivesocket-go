package main

import (
	"bufio"
	"flag"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"github.com/jakewins/reactivesocket-go/pkg/transport/tcp"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	server bool
	host   string
	port   int
	file   string
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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "%%")
		initData := parts[1]
		initMetadata := parts[2]
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

	handler := &rs.RequestHandler{
		HandleRequestResponse: requestResponseInitializer(requestResponseMarbles),
		HandleChannel:         channelHandler(channels),
	}

	log.Fatal(tcp.ListenAndServe(":"+strconv.Itoa(port), func(setup rs.ConnectionSetupPayload, socket rs.ReactiveSocket) (*rs.RequestHandler, error) {
		return handler, nil
	}))
}

func requestResponseInitializer(stuff map[string]map[string]string) func(rs.Payload) rs.Publisher {
	return func(firstPacket rs.Payload) rs.Publisher {
		return nil
	}
}

func channelHandler(channels map[string]map[string][]string) func(rs.Payload, rs.Publisher) rs.Publisher {
	return func(init rs.Payload, in rs.Publisher) rs.Publisher {
		return rs.NewPublisher(func(sub rs.Subscriber) {

		})
	}
}
