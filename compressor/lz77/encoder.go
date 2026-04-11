package lz77

const (
	WindowSize = 32 * 1024

	MaxMatchLength = 258

	MinMatchLength = 3
	hashShift      = 5

	hashSize    = 1 << 15
	hashMask    = hashSize - 1
	maxChainLen = 128
)

func Encode(data []byte) []Token {
	n := len(data)
	if n == 0 {
		return nil
	}

	tokens := make([]Token, 0, n/2)

	head := make([]int, hashSize)

	for i := range head {
		head[i] = -1
	}

	prev := make([]int, n)

	for i := range prev {
		prev[i] = -1
	}

	pos := 0
	for pos < n {
		bestLen := 0
		bestDist := 0

		if pos+MinMatchLength <= n {
			h := hash3(data, pos)

			chainLen := 0
			candidate := head[h]
			windowStart := pos - WindowSize
			if windowStart < 0 {
				windowStart = 0
			}
			for candidate >= windowStart && chainLen < maxChainLen {
				ml := matchLength(data, candidate, pos, n)
				if ml > bestLen {
					bestLen = ml
					bestDist = pos - candidate
					if bestLen == MaxMatchLength {
						break
					}
				}
				chainLen++
				candidate = prev[candidate]
			}

			prev[pos] = head[h]
			head[h] = pos
		}

		if bestLen >= MinMatchLength {
			tokens = append(tokens, NewBackReference(bestDist, bestLen))

			for i := 0; i < bestLen; i++ {
				next := pos + i
				if next+MinMatchLength <= n {
					h := hash3(data, next)
					prev[next] = head[h]
					head[h] = next
				}
			}
			pos += bestLen
		} else {
			tokens = append(tokens, NewLiteral(data[pos]))
			pos++
		}
	}
	return tokens
}

func hash3(data []byte, pos int) int {
	return (int(data[pos])<<(2*hashShift) ^
		int(data[pos+1])<<hashShift) ^
		int(data[pos+2])&hashMask
}

func matchLength(data []byte, a, b, n int) int {
	limit := MaxMatchLength
	if n-b < limit {
		limit = n - b
	}
	if n-a < limit {
		limit = n - a
	}

	length := 0
	for length < limit && data[a+length] == data[b+length] {
		length++
	}
	return length
}
