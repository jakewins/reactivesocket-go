package codec

import "io"

type FrameDecoder struct {
	source *io.Reader
}

func (d *FrameDecoder) Read(target *Frame) error {

}

type FrameEncoder struct {
	sink *io.Writer
}

func (self *FrameEncoder) Write(frame *Frame) error {
	return nil
}

type Frame struct {
	buf []byte
}

func (f *Frame) Data() []byte {
	return nil
}

func (f *Frame) Metadata() []byte {
	return nil
}