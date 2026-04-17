package archive

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type readerEntry struct {
	header     FileHeader
	dataOffset uint64
}

type Reader struct {
	r       *bytes.Reader
	entries []readerEntry
}

var minArchiveSize = len(Magic) + 1 + 2 + 8

func NewReader(data []byte) (*Reader, error) {
	if len(data) < minArchiveSize {
		return nil, fmt.Errorf("archive: data too short (%d bytes, minimum %d)", len(data), minArchiveSize)
	}

	r := bytes.NewReader(data)

	magic := make([]byte, len(Magic))

	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, fmt.Errorf("archive: reading magic: %w", err)
	}

	if !bytes.Equal(magic, Magic) {
		return nil, fmt.Errorf("archive: invalid magic bytes: %x", magic)
	}

	version, err := r.ReadByte()

	if err != nil {
		return nil, fmt.Errorf("archive: reading version: %w", err)
	}
	if version != Version {
		return nil, fmt.Errorf("archive: unsupported version %d (expected %d)", version, Version)
	}

	if _, err := r.Seek(-8, io.SeekEnd); err != nil {
		return nil, fmt.Errorf("archive: seeking to footer: %w", err)
	}

	var centralDirOffset uint64

	if err := binary.Read(r, binary.LittleEndian, &centralDirOffset); err != nil {
		return nil, fmt.Errorf("archive: reading central directory offset: %w", err)
	}

	if _, err := r.Seek(int64(centralDirOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("archive: seeking to central directory: %w", err)
	}

	var entryCount uint16
	if err := binary.Read(r, binary.LittleEndian, &entryCount); err != nil {
		return nil, fmt.Errorf("archive: reading entry count: %w", err)
	}

	entries := make([]readerEntry, 0, entryCount)
	for i := 0; i < int(entryCount); i++ {
		var dataOffset uint64
		if err := binary.Read(r, binary.LittleEndian, &dataOffset); err != nil {
			return nil, fmt.Errorf("archive: reading data offset #%d: %w", i, err)
		}

		var h FileHeader
		if _, err := h.ReadFrom(r); err != nil {
			return nil, fmt.Errorf("archive: reading central header #%d: %w", i, err)
		}

		entries = append(entries, readerEntry{
			header:     h,
			dataOffset: dataOffset,
		})
	}
	return &Reader{
		r:       bytes.NewReader(data),
		entries: entries,
	}, nil
}

func (ar *Reader) Files() []FileHeader {
	headers := make([]FileHeader, len(ar.entries))
	for i, e := range ar.entries {
		headers[i] = e.header
	}
	return headers
}

func (ar *Reader) ReadFile(name string) ([]byte, error) {
	for _, e := range ar.entries {
		if e.header.Name != name {
			continue
		}

		if _, err := ar.r.Seek(int64(e.dataOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("archive: seeking to data for %q: %w", name, err)
		}

		data := make([]byte, e.header.CompressedSize)
		if _, err := io.ReadFull(ar.r, data); err != nil {
			return nil, fmt.Errorf("archive: reading data for %q: %w", name, err)
		}

		return data, nil
	}

	return nil, fmt.Errorf("archive: file %q not found", name)
}
