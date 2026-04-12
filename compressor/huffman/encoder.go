package huffman

import (
	"bytes"
	"encoding/binary"
	"sort"
)

type codeEntry struct {
	sym  byte
	code string
}

func packCodeBits(code string) []byte {
	nBytes := (len(code) + 7) / 8
	buf := make([]byte, nBytes)
	for i := 0; i < len(code); i++ {
		if code[i] == '1' {
			buf[i/8] |= 1 << uint(7-i%8)
		}
	}
	return buf
}

func Encode(data []byte) ([]byte, error) {
	var header bytes.Buffer

	if err := binary.Write(&header, binary.BigEndian, uint32(len(data))); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		if err := binary.Write(&header, binary.BigEndian, uint16(0)); err != nil {
			return nil, err
		}
		return header.Bytes(), nil
	}

	freq := CountFrequencies(data)
	root := BuildTree(freq)
	codes := GenerateCodes(root)

	entries := make([]codeEntry, 0, len(codes))

	for sym, code := range codes {
		entries = append(entries, codeEntry{sym: sym, code: code})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].sym < entries[j].sym
	})

	if err := binary.Write(&header, binary.BigEndian, uint16(len(entries))); err != nil {
		return nil, err
	}

	for _, e := range entries {
		header.WriteByte(e.sym)
		header.WriteByte(byte(len(e.code)))
		header.Write(packCodeBits(e.code))
	}

	var encoded bytes.Buffer
	bw := NewBitWriter(&encoded)

	for _, b := range data {
		code := codes[b]
		for i := 0; i < len(code); i++ {
			bit := 0
			if code[i] == '1' {
				bit = 1
			}
			if err := bw.WriteBit(bit); err != nil {
				return nil, err
			}
		}
	}
	if err := bw.Flush(); err != nil {
		return nil, err
	}
	header.Write(encoded.Bytes())
	return header.Bytes(), nil
}
