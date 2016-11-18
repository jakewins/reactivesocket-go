package response

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

func IsCompleteStream(f *frame.Frame) bool {
	return f.Flags()&header.FlagResponseComplete != 0
}
