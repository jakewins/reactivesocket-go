package keepalive

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
)

func Encode(bufPtr *[]byte, respond bool) {
	var flags uint16
	buf := header.ResizeSlice(bufPtr, header.ComputeLength(0, 0))

	if respond {
		flags |= header.FlagKeepaliveRespond
	}
	header.EncodeHeader(buf, flags, header.FTKeepAlive, 0)
}

func Describe(buf []byte) string {
	if header.Flags(buf)&header.FlagKeepaliveRespond != 0 {
		return "KeepAlive{respond=yes}"
	}
	return "KeepAlive{respond=no}"
}
