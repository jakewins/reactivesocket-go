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
