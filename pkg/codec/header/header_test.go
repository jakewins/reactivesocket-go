package header_test

import (
	"testing"
	"github.com/jakewins/reactivesocket-go/pkg/codec/header"
)

var capacities = [][]int{
	// Original, Ensure, Expected
	{         0,      0,       0},
	{         1,      1,       1},
	{         1,    512,     512},
	{       512,    512,     512},
	{       512,    513,    1024},
}

func TestEnsureCapacity(t *testing.T) {
	for sampleNo, test := range capacities {
		original, ensure, expected := test[0], test[1], test[2]
		slice := make([]byte, original)

		header.EnsureCapacity(&slice, ensure)

		if cap(slice) != expected {
			t.Errorf("Sample %d: Expected slice to have been replaced by a %s-capacity one, found %d",
				sampleNo, expected, cap(slice))
		}
	}
}