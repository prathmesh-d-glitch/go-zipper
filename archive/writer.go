package archive

import (
	"bytes"
	"encoding/binary"
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

func (w *Writer) Close() ([]byte, error) {
	if w.closed {
		return nil, fmt.Errorf("archive: writer is already closed")
	}

	w.closed = true

	centralDirOffset := uint64(w.written)

	if err := binary.Write(&w.buf, binary.LittleEndian, uint16(len(w.entries))); err != nil {
		return nil, fmt.Errorf("archive: writing entry count: %w", err)
	}

	for _, e := range w.entries {
		if err := binary.Write(&w.buf, binary.LittleEndian, e.dataOffset); err != nil {
			return nil, fmt.Errorf("archive: writing data offset: %w", err)
		}
		if _, err := e.header.WriteTo(&w.buf); err != nil {
			return nil, fmt.Errorf("archive: writing central header: %w", err)
		}
	}

	if err := binary.Write(&w.buf, binary.LittleEndian, centralDirOffset); err != nil {
		return nil, fmt.Errorf("archive: writing central directory offset: %w", err)
	}

	return w.buf.Bytes(), nil
}
