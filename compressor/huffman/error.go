package huffman

import "errors"

var (
	ErrInvalidBit = errors.New("bit value must be either 0 or 1")

	ErrInvalidBitCount = errors.New("bit count must be between 1 and 32")
)
