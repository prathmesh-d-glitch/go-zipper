package lz77

import "fmt"

func Decode(tokens []Token) ([]byte, error) {
	out := make([]byte, 0, len(tokens)*2)

	for i, tok := range tokens {
		if tok.isLiteral {
			out = append(out, tok.Literal)
			continue
		}

		if tok.Offset <= 0 || tok.Offset > len(out) {
			return nil, fmt.Errorf(
				"lz77: token %d: invalid distance %d (output length %d)",
				i, tok.Offset, len(out),
			)
		}

		if tok.Length <= 0 {
			return nil, fmt.Errorf(
				"lz77: token %d: invalid length %d",
				i, tok.Length,
			)
		}

		start := len(out) - tok.Offset
		for j := 0; j < tok.Length; j++ {
			out = append(out, out[start+j])
		}
	}

	return out, nil
}
