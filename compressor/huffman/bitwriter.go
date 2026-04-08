package huffman

import "io"

type BitWriter struct {
	w           io.Writer
	buf         byte
	bitCount    uint8
	paddingBits uint8
}

func NewBitWriter(w io.Writer) *BitWriter {
	return &BitWriter{w: w}
}
