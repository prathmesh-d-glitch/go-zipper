package lz77

import "fmt"

type Token struct {
	isLiteral bool
	Literal   byte
	Offset    int
	Length    int
}

func NewLiteral(b byte) Token {
	return Token{isLiteral: true, Literal: b}
}

func NewBackReference(offset, length int) Token {
	return Token{isLiteral: false, Offset: offset, Length: length}
}

//for debugging
func (t *Token) String() string {
	if t.isLiteral {
		if t.Literal >= 0x20 && t.Literal <= 0x7E {
			return fmt.Sprintf("Literal('%c')", t.Literal)
		}
		return fmt.Sprintf("Literal(0x%02X)", t.Literal)
	}
	return fmt.Sprintf("BackRef(d=%d, l=%d)", t.Offset, t.Length)
}
