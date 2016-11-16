package proto_test

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/request"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
	"github.com/jakewins/reactivesocket-go/pkg/internal/proto"
	"github.com/jakewins/reactivesocket-go/pkg/rs"
	"testing"
)

func TestRequestWithInitialNLeadsToRequestN(t *testing.T) {
	r := recorder{}
	p := proto.NewProtocol(&rs.RequestHandler{
		HandleChannel: channelFactory(wall, sequencer(1, 10)),
	}, 0, r.Record)

	p.HandleFrame(frame.RequestWithInitialN(1337, 2, request.RequestFlagRequestChannelN, header.FTRequestChannel, nil, nil))

	if err := r.AssertRecorded([]*frame.Frame{
		frame.Response(1337, 0, nil, []byte{0, 0, 0, 1}),
		frame.Response(1337, 0, nil, []byte{0, 0, 0, 2}),
	}); err != nil {
		t.Error(err)
	}
}
