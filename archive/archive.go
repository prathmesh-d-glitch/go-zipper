package archive

import (
	"fmt"

	"github.com/prathmesh-d-glitch/go-zipper/compressor/huffman"
	"github.com/prathmesh-d-glitch/go-zipper/compressor/lz77"
	"github.com/prathmesh-d-glitch/go-zipper/utils"
)

func CompressFiles(paths []string) ([]byte, error) {
	w := NewWriter()

	for _, path := range paths {
		data, err := utils.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("archive: reading %q: %w", path, err)
		}

		crc := utils.Checksum(data)

		tokens := lz77.Encode(data)
		serialised := lz77.SerializeTokens(tokens)
		compressed, err := huffman.Encode(serialised)

		if err != nil {
			return nil, fmt.Errorf("archive: compressing %q: %w", path, err)
		}

		name := utils.GetFileName(path)

		if err := w.AddFile(name, compressed, uint64(len(data)), crc); err != nil {
			return nil, err
		}

	}

	return w.Close()
}
