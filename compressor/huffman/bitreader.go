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

func (br *BitReader) ReadBit() (int, error) {
	if br.bitCount == 0 {
		var b [1]byte
		if _, err := io.ReadFull(br.r, b[:]); err != nil {
			return 0, err
		}
		br.buf = b[0]
		br.bitCount = 8
	}
	br.bitCount--
	bit := int((br.buf >> br.bitCount) & 1)

	return bit, nil
}

func (br *BitReader) ReadByte() (byte, error) {
	var b byte
	for i := 7; i >= 0; i-- {
		bit, err := br.ReadBit()
		if err != nil {
			return 0, err
		}
		b |= byte(bit) << i
	}
	return b, nil
}

func (br *BitReader) ReadBits(n int) (uint32, error) {
	if n < 1 || n > 32 {
		return 0, ErrInvalidBitCount
	}

	var value uint32
	for i := 0; i < n; i++ {
		bit, err := br.ReadBit()
		if err != nil {
			return 0, err
		}
		value = (value << 1) | uint32(bit)
	}

	return value, nil
}
