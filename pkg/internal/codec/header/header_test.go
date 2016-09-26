package header_test

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/header"
	"testing"
)

var capacities = [][]int{
	// Original, Ensure, Expected
	{0, 0, 0},
	{1, 1, 1},
	{1, 2, 512},
	{1, 512, 512},
	{512, 512, 512},
	{512, 513, 1024},
}

func TestEnsureCapacity(t *testing.T) {
	for sampleNo, test := range capacities {
		original, ensure, expected := test[0], test[1], test[2]
		slice := make([]byte, original)

		header.ResizeSlice(&slice, ensure)

		if cap(slice) != expected {
			t.Errorf("Sample %d: Expected slice to have been replaced by a %s-capacity one, found %d",
				sampleNo, expected, cap(slice))
		}
		if len(slice) != ensure {
			t.Errorf("Sample %d: Expected len() to be %d, got %d", sampleNo, ensure, len(slice))
		}
	}
}
