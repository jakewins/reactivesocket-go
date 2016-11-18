package rs

type Payload interface {
	Metadata() []byte
	Data() []byte
}

type ReactiveSocket interface {
	FireAndForget(Payload) Publisher
	RequestResponse(Payload) Publisher
	RequestStream(Payload) Publisher
	RequestSubscription(Payload) Publisher
	RequestChannel(Publisher) Publisher
	//MetadataPush(Payload) Publisher
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
		data = make([]byte, len(sourceData))
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

func NewSetupPayload(metadataMimeType, dataMimeType string, metadata, data []byte) ConnectionSetupPayload {
	return &anonymousSetupPayload{
		anonymousPayload{metadata, data},
		metadataMimeType,
		dataMimeType,
	}
}

type anonymousSetupPayload struct {
	anonymousPayload
	metadataMimeType string
	dataMimeType     string
}

func (ap *anonymousSetupPayload) MetadataMimeType() string {
	return ap.metadataMimeType
}
func (ap *anonymousSetupPayload) DataMimeType() string {
	return ap.dataMimeType
}
