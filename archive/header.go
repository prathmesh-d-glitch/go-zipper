package archive

import (
	"encoding/binary"
	"fmt"
	"io"
)

var Magic = []byte{'F', 'Z', 'P'}

const Version byte = 0x01

type FileHeader struct {
	Name           string
	OriginalSize   uint64
	CompressedSize uint64
	CRC32          uint32
}

const headerFixedSize = 2 + 8 + 8 + 4

func (h *FileHeader) Size() int64 {
	return int64(headerFixedSize + len(h.Name))
}

func (h *FileHeader) WriteTo(w io.Writer) (int64, error) {
	var written int64

	nameLen := uint16(len(h.Name))
	if err := binary.Write(w, binary.LittleEndian, nameLen); err != nil {
		return written, fmt.Errorf("archive: writing name length: %w", err)
	}
	written += 2

	n, err := io.WriteString(w, h.Name)
	if err != nil {
		return written, fmt.Errorf("archive: writing name: %w", err)
	}

	written += int64(n)

	if err := binary.Write(w, binary.LittleEndian, h.OriginalSize); err != nil {
		return written, fmt.Errorf("archive: writing original size: %w", err)
	}

	written += 8

	if err := binary.Write(w, binary.LittleEndian, h.CompressedSize); err != nil {
		return written, fmt.Errorf("archive: writing compressed size: %w", err)
	}
	written += 8

	if err := binary.Write(w, binary.LittleEndian, h.CRC32); err != nil {
		return written, fmt.Errorf("archive: writing CRC32: %w", err)
	}
	written += 4

	return written, nil

}

func (h *FileHeader) ReadFrom(r io.Reader) (int64, error) {
	var read int64

	var nameLen uint16
	if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
		return read, fmt.Errorf("archive: reading name length: %w", err)
	}
	read += 2

	name := make([]byte, nameLen)
	if _, err := io.ReadFull(r, name); err != nil {
		return read, fmt.Errorf("archive: reading name: %w", err)
	}
	h.Name = string(name)
	read += int64(nameLen)

	if err := binary.Read(r, binary.LittleEndian, &h.OriginalSize); err != nil {
		return read, fmt.Errorf("archive: reading original size: %w", err)
	}
	read += 8

	if err := binary.Read(r, binary.LittleEndian, &h.CompressedSize); err != nil {
		return read, fmt.Errorf("archive: reading compressed size: %w", err)
	}
	read += 8

	if err := binary.Read(r, binary.LittleEndian, &h.CRC32); err != nil {
		return read, fmt.Errorf("archive: reading CRC32: %w", err)
	}
	read += 4

	return read, nil
}
