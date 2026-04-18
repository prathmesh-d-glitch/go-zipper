package lz77

import "fmt"

const (
	tokenTypeLiteral byte = 0x00
	tokenTypeBackRef byte = 0x01
)

func SerializeTokens(tokens []Token) []byte {
	// Pre-allocate assuming mostly literals (2 bytes each).
	buf := make([]byte, 0, len(tokens)*2)

	for _, t := range tokens {
		if t.isLiteral {
			buf = append(buf, tokenTypeLiteral, t.Literal)
		} else {
			buf = append(buf, tokenTypeBackRef)
			buf = append(buf, byte(t.Offset), byte(t.Offset>>8))
			buf = append(buf, byte(t.Length), byte(t.Length>>8))
		}
	}
	return buf
}

func DeserializeTokens(data []byte) ([]Token, error) {
	tokens := make([]Token, 0, len(data)/2)
	i := 0
	for i < len(data) {
		switch data[i] {
		case tokenTypeLiteral:
			if i+1 >= len(data) {
				return nil, fmt.Errorf("lz77: truncated literal token at offset %d", i)
			}
			tokens = append(tokens, NewLiteral(data[i+1]))
			i += 2

		case tokenTypeBackRef:
			if i+4 >= len(data) {
				return nil, fmt.Errorf("lz77: truncated back-reference token at offset %d", i)
			}
			distance := int(data[i+1]) | int(data[i+2])<<8
			length := int(data[i+3]) | int(data[i+4])<<8
			tokens = append(tokens, NewBackReference(distance, length))
			i += 5

		default:
			return nil, fmt.Errorf("lz77: unknown token type 0x%02X at offset %d", data[i], i)
		}
	}
	return tokens, nil
}
