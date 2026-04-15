package archive

import (
	"bytes"
	"fmt"
)

type centralEntry struct {
	header     FileHeader
	dataOffset uint64
}

type Writer struct {
	buf     bytes.Buffer
	entries []centralEntry
	written int64
	closed  bool
}

func NewWriter() *Writer {
	w := &Writer{}
	w.buf.Write(Magic)
	w.buf.WriteByte(Version)
	w.written = int64(len(Magic)) + 1
	return w
}

func (w *Writer) AddFile(name string, compressedData []byte, originalSize uint64, crc32 uint32) error {
	if w.closed {
		return fmt.Errorf("archive: writer is closed")
	}

	h := FileHeader{
		Name:           name,
		OriginalSize:   originalSize,
		CompressedSize: uint64(len(compressedData)),
		CRC32:          crc32,
	}

	n, err := h.WriteTo(&w.buf)

	if err != nil {
		return fmt.Errorf("archive: writing header for %q: %w", name, err)
	}

	w.written += n

	dataOffset := uint64(w.written)

	nn, err := w.buf.Write(compressedData)

	if err != nil {
		return fmt.Errorf("archive: writing data for %q: %w", name, err)
	}

	w.written += int64(nn)

	w.entries = append(w.entries, centralEntry{
		header:     h,
		dataOffset: dataOffset,
	})

	return nil
}
