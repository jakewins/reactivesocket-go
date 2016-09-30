package rs

type Payload interface {
	Metadata() []byte
	Data() []byte
}

type ReactiveSocket interface {
}

func NewPayload(metadata, data []byte) Payload {
	return &anonymousPayload{metadata, data}
}
func CopyPayload(p Payload) Payload {
	var meta []byte
	var data []byte

	sourceMeta, sourceData := p.Metadata(), p.Data()
	if sourceMeta != nil {
		meta = make([]byte, len(sourceMeta))
		copy(meta, sourceMeta)
	}
	if sourceData != nil {
		data := make([]byte, len(sourceData))
		copy(data, sourceData)
	}
	return &anonymousPayload{meta, data}
}

type anonymousPayload struct {
	metadata []byte
	data     []byte
}

func (ap *anonymousPayload) Metadata() []byte {
	return ap.metadata
}
func (ap *anonymousPayload) Data() []byte {
	return ap.data
}
