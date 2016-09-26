package rs

type Payload interface {
	Metadata() []byte
	Data() []byte
}

type ReactiveSocket interface {
}
