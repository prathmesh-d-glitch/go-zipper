package huffman

import "io"

type BitReader struct {
	r        io.Reader
	buf      byte
	bitCount uint8
}

func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{r: r}
}
