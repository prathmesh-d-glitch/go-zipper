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

func DecompressArchive(data []byte, destDir string) error {
	reader, err := NewReader(data)

	if err != nil {
		return err
	}

	for _, hdr := range reader.Files() {
		compressed, err := reader.ReadFile(hdr.Name)
		if err != nil {
			return err
		}

		serialised, err := huffman.Decode(compressed)
		if err != nil {
			return fmt.Errorf("archive: huffman decode %q %w", hdr.Name, err)
		}

		tokens, err := lz77.DeserializeTokens(serialised)
		if err != nil {
			return fmt.Errorf("archive: token deserialise %q: %w", hdr.Name, err)
		}

		original, err := lz77.Decode(tokens)
		if err != nil {
			return fmt.Errorf("archive: lz77 decode %q: %w", hdr.Name, err)
		}

		if !utils.Verify(original, hdr.CRC32) {
			return fmt.Errorf(
				"archive: CRC32 mismatch for %q: got 0x%08X, want 0x%08X",
				hdr.Name, utils.Checksum(original), hdr.CRC32,
			)
		}

		if uint64(len(original)) != hdr.OriginalSize {
			return fmt.Errorf(
				"archive: size mismatch for %q,got %d, want %d",
				hdr.Name, len(original), hdr.OriginalSize,
			)
		}

		outPath := utils.GetFileName(hdr.Name)
		if err := utils.WriteFile(destDir+"/"+outPath, original); err != nil {
			return fmt.Errorf("archive: writing %q: %w", outPath, err)
		}
	}
	return nil
}
