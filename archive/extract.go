package archive

import (
	"fmt"

	"github.com/prathmesh-d-glitch/go-zipper/compressor/huffman"
	"github.com/prathmesh-d-glitch/go-zipper/compressor/lz77"
	"github.com/prathmesh-d-glitch/go-zipper/utils"
)

type ExtractedFile struct {
	Name string
	Data []byte
}

func ExtractFiles(data []byte) ([]ExtractedFile, error) {
	reader, err := NewReader(data)
	if err != nil {
		return nil, err
	}

	var out []ExtractedFile

	for _, hdr := range reader.Files() {
		compressed, err := reader.ReadFile(hdr.Name)
		if err != nil {
			return nil, err
		}

		serialised, err := huffman.Decode(compressed)
		if err != nil {
			return nil, fmt.Errorf("archive: huffman decode %q: %w", hdr.Name, err)
		}

		tokens, err := lz77.DeserializeTokens(serialised)
		if err != nil {
			return nil, fmt.Errorf("archive: token deserialise %q: %w", hdr.Name, err)
		}

		original, err := lz77.Decode(tokens)
		if err != nil {
			return nil, fmt.Errorf("archive: lz77 decode %q: %w", hdr.Name, err)
		}

		if !utils.Verify(original, hdr.CRC32) {
			return nil, fmt.Errorf(
				"archive: CRC32 mismatch for %q: got 0x%08X, want 0x%08X",
				hdr.Name, utils.Checksum(original), hdr.CRC32,
			)
		}

		if uint64(len(original)) != hdr.OriginalSize {
			return nil, fmt.Errorf(
				"archive: size mismatch for %q: got %d, want %d",
				hdr.Name, len(original), hdr.OriginalSize,
			)
		}

		out = append(out, ExtractedFile{Name: hdr.Name, Data: original})
	}

	return out, nil
}
