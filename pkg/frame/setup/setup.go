package setup


import (
	"github.com/jakewins/reactivesocket-go/pkg/codec/setup"
	"github.com/jakewins/reactivesocket-go/pkg/frame"
)

const (
	FlagWillHonorLease = setup.SetupFlagWillHonorLease
	FlagStrictInterpretation = setup.SetupFlagStrictInterpretation
)

func Encode(target *frame.Frame, flags uint16, keepaliveInterval, maxLifetime uint32,
                metadataMimeType, dataMimeType string, metadata, data []byte) {
	setup.Encode(&target.Buf, flags, keepaliveInterval, maxLifetime, metadataMimeType,
		dataMimeType, metadata, data)
}

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