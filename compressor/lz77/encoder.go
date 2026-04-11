package lz77

const (
	WindowSize = 32 * 1024

	MaxMatchLength = 258

	MinMatchLength = 3
	hashShift      = 5

	hashSize = 1 << 15
	hashMask = hashSize - 1
)

func hash3(data []byte, pos int) int {
	return (int(data[pos])<<(2*hashShift) ^
		int(data[pos+1])<<hashShift) ^
		int(data[pos+2])&hashMask
}
