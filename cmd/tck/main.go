package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"github.com/jakewins/reactivesocket-go/pkg/transport/tcp"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
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

func channelHandler(channels map[string]map[string][]string) func(rs.Publisher) rs.Publisher {
	return func(in rs.Publisher) rs.Publisher {
		sub := NewPuppetSubscriber()
		pub := NewPuppetPublisher()
		in.Subscribe(sub)

		go channelWorker(channels, sub, pub)

		return pub
	}
}

func channelWorker(channels map[string]map[string][]string, in *puppetSubscriber, out *puppetPublisher) {
	// First inbound message used to determine how to behave
	in.request(1)
	init, err := in.awaitNext()
	if err != nil {
		panic(err.Error())
	}
	script := channels[string(init.Data())][string(init.Metadata())]

	for _, command := range script {
		args := strings.Split(command, "%%")
		//fmt.Printf("%v\n", args)
		switch args[0] {
		case "respond":
			out.publish(rs.NewPayload(nil, []byte(args[1])))
		case "request":
			in.request(parseInt(args[1]))
		case "await":
			switch args[1] {
			case "atLeast":
				if err := in.awaitAtLeast(parseInt(args[3])); err != nil {
					panic(err.Error())
				}
			case "terminal":
				if err := in.awaitTerminal(); err != nil {
					panic(err.Error())
				}
			default:
				panic(fmt.Sprintf("Unknown TCK command for channel: %v", args))
			}
		case "assert":
			switch args[1] {
			case "completed":
				if err := in.assertComplete(); err != nil {
					panic(err.Error())
				}
			}
		default:
			panic(fmt.Sprintf("Unknown TCK command for channel: %v", args))
		}
	}
}

func NewPuppetSubscriber() *puppetSubscriber {
	return &puppetSubscriber{
		inbound: make(chan rs.Payload, 16),
		control: make(chan string, 2),
	}
}

type puppetSubscriber struct {
	subscription rs.Subscription
	inbound      chan rs.Payload
	control      chan string
	received     []rs.Payload
	state        string
}

func (p *puppetSubscriber) OnSubscribe(s rs.Subscription) {
	p.subscription = s
}
func (p *puppetSubscriber) OnNext(v rs.Payload) {
	p.inbound <- rs.CopyPayload(v)
}
func (p *puppetSubscriber) OnError(err error) {
	p.state = "ERROR"
	p.control <- fmt.Sprintf("ERROR: %s", err.Error())
}
func (p *puppetSubscriber) OnComplete() {
	p.state = "COMPLETE"
	p.control <- "COMPLETE"
}
func (p *puppetSubscriber) request(n int) {
	p.subscription.Request(n)
}
func (p *puppetSubscriber) awaitAtLeast(n int) error {
	for {
		if len(p.received) >= n {
			return nil
		}
		select {
		case msg := <-p.inbound:
			p.received = append(p.received, msg)
		case <-time.After(time.Second * 5):
			return fmt.Errorf("Expected %d messages, timed out before that.", n)
		}
	}
}
func (p *puppetSubscriber) awaitTerminal() error {
	select {
	case <-p.control:
		return nil
	case <-time.After(time.Second * 5):
		return fmt.Errorf("Expected subscriber to be completed, timed out before that.")
	}
}
func (p *puppetSubscriber) awaitNext() (rs.Payload, error) {
	nextIndex := len(p.received)
	if err := p.awaitAtLeast(1); err != nil {
		return nil, err
	}
	return p.received[nextIndex], nil
}
func (p *puppetSubscriber) assertComplete() (error) {
	if p.state != "COMPLETE" {
		return fmt.Errorf("Expected stream to be COMPLETE, found %s", p.state)
	}
	return nil
}

func NewPuppetPublisher() *puppetPublisher {
	return &puppetPublisher{}
}

type puppetPublisher struct {
	subscriber rs.Subscriber
	requested  int32
	control    chan string
}

func (p *puppetPublisher) Subscribe(s rs.Subscriber) {
	p.subscriber = s
	s.OnSubscribe(rs.NewSubscription(
		func(n int) {
			atomic.AddInt32(&p.requested, int32(n))
		},
		func() {
			// Cancelled
		},
	))
}
func (p *puppetPublisher) publish(v rs.Payload) {
	for {
		if atomic.LoadInt32(&p.requested) > 0 {
			atomic.AddInt32(&p.requested, -1)
			break
		}
		time.Sleep(time.Millisecond * 5)
	}
	p.subscriber.OnNext(v)
}

func parseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return v
}
