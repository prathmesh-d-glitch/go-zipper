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

func (bw *BitWriter) WriteBit(bit int) error {
	if bit != 0 && bit != 1 {
		return ErrInvalidBit
	}

	if bit == 1 {
		bw.buf |= 1 << (7 - bw.bitCount)
	}

	if bw.bitCount == 8 {
		if _, err := bw.w.Write([]byte{bw.buf}); err != nil {
			return err
		}
		bw.buf = 0
		bw.bitCount = 0
	}

	return nil
}

func (bw *BitWriter) WriteBytes(b byte) error {
	for i := 7; i > -0; i-- {
		if err := bw.WriteBit(int((b >> i) & 1)); err != nil {
			return err
		}
	}
	return nil
}

func (bw *BitWriter) WriteBits(value uint32, n int) error {
	if n < 1 || n > 32 {
		return ErrInvalidBitCount
	}

	for i := n - 1; i >= 0; i-- {
		if err := bw.WriteBit(int((value >> i) & 1)); err != nil {
			return err
		}
	}
	return nil
}

func (bw *BitWriter) Flush() error {
	if bw.bitCount == 0 {
		return nil
	}

	bw.paddingBits = 8 - bw.bitCount
	if _, err := bw.w.Write([]byte{bw.buf}); err != nil {
		return err
	}
	bw.buf = 0
	bw.bitCount = 0

	return nil
}

func (bw *BitWriter) PaddingBits() uint8 {
	return bw.paddingBits
}
