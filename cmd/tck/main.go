package main

import (
	"bufio"
	"encoding/json"
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
	client bool
	host   string
	port   int
	file   string
)

func init() {
	flag.BoolVar(&server, "server", false, "To launch the server")
	flag.BoolVar(&client, "client", false, "To launch the client")
	flag.StringVar(&host, "host", "localhost", "For the client only, determine host to connect to")
	flag.IntVar(&port, "port", 4567, "For client, port to connect to. For server, port to bind to")
	flag.StringVar(&file, "file", "", "Path to script file to run")
}

func main() {
	flag.Parse()

	if server {
		runServer(port, file)
	} else if client {
		runClient(host, port, file)
	}
}

func runClient(host string, port int, path string) {
	address := host + ":" + strconv.Itoa(port)

	fmt.Printf("[TCK] Client started, playing script against [%s]\n", address)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var scripts [][]string
	var current []string
	for scanner.Scan() {
		switch scanner.Text() {
		case "!":
			if current != nil {
				scripts = append(scripts, current)
			}
			current = nil
		default:
			current = append(current, scanner.Text())
		}
	}
	scripts = append(scripts, current)

	subscriptions := make(map[string]*puppetSubscriber)

	for _, script := range scripts {
		shouldPass := true
		name := "Unknown"
		for i := 0; i < len(script); i++ {
			line := script[i]
			parts := strings.Split(line, "%%")
			fmt.Printf("% v\n", parts)
			switch parts[0] {
			case "name":
				name = parts[1]
				fmt.Printf("[TCK] Starting %s\n", name)
			case "pass":
				shouldPass = true
			case "fail":
				shouldPass = false
			case "channel":
				end := i
				for ; end < len(script); end++ {
					line = script[end]
					if line == "}" {
						break
					}
				}
				outcome := clientChannel(address, script[i:end])
				i = end
				if shouldPass && outcome != nil {
					panic(outcome)
				} else if !shouldPass && outcome == nil {
					panic(fmt.Errorf("Expected %s to fail, but it passed.", name))
				}
			case "subscribe":
				sub := NewPuppetSubscriber()
				subscriptions[parts[2]] = sub
				socket, err := tcp.Dial(address, rs.NewSetupPayload("", "", nil, nil))
				payload := rs.NewPayload([]byte(parts[3]), []byte(parts[4]))
				if err != nil {
					panic(err)
				}
				switch parts[1] {
				case "fnf":
					socket.FireAndForget(payload).Subscribe(sub)
				case "sub":
					socket.RequestSubscription(payload).Subscribe(sub)
				case "rs":
					socket.RequestStream(payload).Subscribe(sub)
				case "rr":
					socket.RequestResponse(payload).Subscribe(sub)
				default:
					panic(fmt.Sprintf("Unknown client action: %v %v", parts, shouldPass))
				}
			case "request":
				n, err := strconv.Atoi(parts[1])
				if err != nil {
					panic(err)
				}
				subscriptions[parts[2]].request(n)
			case "cancel":
				subscriptions[parts[1]].cancel()
			case "await":
				switch parts[1] {
				case "terminal":
					subscriptions[parts[2]].awaitTerminal()
				}
			case "assert":
				switch parts[1] {
				case "no_error":
					subscriptions[parts[2]].assertComplete()
				}
			case "EOF":
				return
			default:
				panic(fmt.Sprintf("Unknown client action: %v %v", parts, shouldPass))
			}
		}
	}
}

func clientChannel(address string, script []string) error {
	socket, err := tcp.Dial(address, rs.NewSetupPayload("", "", nil, nil))
	if err != nil {
		return err
	}

	pub := NewPuppetPublisher()
	sub := NewPuppetSubscriber()

	first := strings.Split(script[0], "%%")

	socket.RequestChannel(pub).Subscribe(sub)
	sub.request(1)
	pub.publish(rs.NewPayload([]byte(first[2]), []byte(first[1])))

	return runChannelScript("..", script[1:], sub, pub)
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

	var server tcp.Server
	handler := &rs.RequestHandler{
		HandleRequestResponse:     requestResponseHandler(requestResponseMarbles),
		HandleChannel:             channelHandler(channels),
		HandleFireAndForget:       fireAndForgetHandler(&server),
		HandleRequestStream:       streamHandler(requestStreamMarbles),
		HandleRequestSubscription: subscriptionHandler(requestSubscriptionMarbles),
	}

	address := ":" + strconv.Itoa(port)
	server, err = tcp.Listen(address, func(setup rs.ConnectionSetupPayload, socket rs.ReactiveSocket) (*rs.RequestHandler, error) {
		return handler, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("[TCK] Server started, listening on [%s]\n", address)
	log.Fatal(server.Serve())
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
func fireAndForgetHandler(server *tcp.Server) func(rs.Payload) {
	return func(p rs.Payload) {
		if string(p.Data()) == "shutdown" {
			(*server).Shutdown()
		}
	}
}

func requestResponseHandler(scripts map[string]map[string]string) func(rs.Payload) rs.Publisher {
	return subscriptionHandler(scripts)
}
func streamHandler(scripts map[string]map[string]string) func(rs.Payload) rs.Publisher {
	return subscriptionHandler(scripts)
}
func subscriptionHandler(scripts map[string]map[string]string) func(rs.Payload) rs.Publisher {
	return func(init rs.Payload) rs.Publisher {
		script := scripts[string(init.Data())][string(init.Metadata())]
		pub := NewPuppetPublisher()

		var marbleWork = make(chan string, 2)
		go marbleWorker(marbleWork, pub)
		marbleWork <- script
		marbleWork <- "EOF"

		return pub
	}
}

func channelWorker(channels map[string]map[string][]string, in *puppetSubscriber, out *puppetPublisher) {
	// First inbound message used to determine how to behave
	in.request(1)
	init, err := in.awaitNext()
	if err != nil {
		panic(err)
	}
	key := fmt.Sprintf("%s:%s", string(init.Data()), string(init.Metadata()))
	script := channels[string(init.Data())][string(init.Metadata())]

	err = runChannelScript(key, script, in, out)
	if err != nil {
		panic(err)
	}
}

func runChannelScript(key string, script []string, in *puppetSubscriber, out *puppetPublisher) error {
	fmt.Printf("[%s] Starting worker\n", key)

	var marbleWork = make(chan string, 2)
	defer func() { marbleWork <- "EOF" }()
	go marbleWorker(marbleWork, out)

	for _, command := range script {
		args := strings.Split(command, "%%")
		switch args[0] {
		case "respond":
			marbleWork <- args[1]
		case "request":
			in.request(parseInt(args[1]))
		case "await":
			switch args[1] {
			case "atLeast":
				if err := in.awaitAtLeast(parseInt(args[3])); err != nil {
					return fmt.Errorf("[%s] %s", key, err.Error())
				}
			case "terminal":
				if err := in.awaitTerminal(); err != nil {
					return fmt.Errorf("[%s] %s", key, err.Error())
				}
			default:
				return fmt.Errorf("[%s] Unknown TCK command for channel: %v", key, args)
			}
		case "assert":
			switch args[1] {
			case "completed":
				if err := in.assertComplete(); err != nil {
					return fmt.Errorf("[%s] %s", key, err.Error())
				}
			}
		default:
			return fmt.Errorf("[%s] Unknown TCK command for channel: %v", key, args)
		}
	}

	return nil
}

func marbleWorker(work chan string, out *puppetPublisher) {
	var marble string
	for {
		marble = <-work
		if marble == "EOF" {
			return
		}
		var payload = func(key string) rs.Payload {
			return rs.NewPayload([]byte(key), []byte(key))
		}
		if strings.Contains(marble, "&&") {
			parts := strings.Split(marble, "&&")
			marble = parts[0]
			var args map[string]map[string]string
			if err := json.Unmarshal([]byte(parts[1]), &args); err != nil {
				panic(err)
			}
			payload = func(key string) rs.Payload {
				var mappedPayload = args[key]
				if mappedPayload == nil {
					return rs.NewPayload([]byte(key), []byte(key))
				}
				for k, v := range mappedPayload {
					return rs.NewPayload([]byte(v), []byte(k))
				}
				panic("Should never reach here.")
			}
			fmt.Printf("Playing marble `%s` %v\n", marble, args)
		} else {
			fmt.Printf("Playing marble `%s`\n", marble)
		}

		for _, c := range marble {
			switch string(c) {
			case "-": // do nothing
			case "|": // completed stream
				out.complete()
			case "#": // error
				out.causeError()
			default:
				out.publish(payload(string(c)))
			}
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
func (p *puppetSubscriber) cancel() {
	p.subscription.Cancel()
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
func (p *puppetSubscriber) assertComplete() error {
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
func (p *puppetPublisher) complete() {
	p.subscriber.OnComplete()
}
func (p *puppetPublisher) causeError() {
	p.subscriber.OnError(fmt.Errorf("Intentional error induced by TCK script."))
}
func parseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return v
}
