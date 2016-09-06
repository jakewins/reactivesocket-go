package setup_test

import (
	"testing"
	"github.com/jakewins/reactivesocket-go/pkg/codec/setup"
)

func TestSetupFrameEncoding(t *testing.T) {
	_ := setup.NewSetupFrame(0, 1, 2, "test/test+meta", "test/test+data", []byte{1,2,3}, []{4,5,6})
}

