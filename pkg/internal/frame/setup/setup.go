package setup

import (
	"github.com/jakewins/reactivesocket-go/pkg/internal/codec/setup"
	"github.com/jakewins/reactivesocket-go/pkg/internal/frame"
)

const (
	FlagWillHonorLease       = setup.SetupFlagWillHonorLease
	FlagStrictInterpretation = setup.SetupFlagStrictInterpretation
)

func Flags(f *frame.Frame) uint16 {
	return setup.Flags(f.Buf)
}

func Version(f *frame.Frame) uint32 {
	return setup.Version(f.Buf)
}

func KeepaliveInterval(f *frame.Frame) uint32 {
	return setup.KeepaliveInterval(f.Buf)
}

func MaxLifetime(f *frame.Frame) uint32 {
	return setup.MaxLifetime(f.Buf)
}

func MetadataMimeType(f *frame.Frame) string {
	return setup.MetadataMimeType(f.Buf)
}

func DataMimeType(f *frame.Frame) string {
	return setup.DataMimeType(f.Buf)
}
