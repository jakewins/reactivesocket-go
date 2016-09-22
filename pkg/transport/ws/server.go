package ws

import "golang.org/x/net/websocket"

var ReactiveSocketCodec = websocket.Codec{marshal, unmarshal}

func marshal(v interface{}) (data []byte, payloadType byte, err error) {
	return nil, 0, nil
}

func unmarshal(data []byte, payloadType byte, v interface{}) (err error) {
	return nil, 0, nil
}

func asd() {
	// I want to create a new server..

}