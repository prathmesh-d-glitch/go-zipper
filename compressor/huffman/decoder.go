package huffman

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const headerMinSize = 6

func Decode(data []byte) ([]byte, error) {
	if len(data) < headerMinSize {
		return nil, fmt.Errorf("huffman: input too short (%d bytes); minimum is %d", len(data), headerMinSize)
	}

	r := bytes.NewReader(data)

	var originalLen uint32
	if err := binary.Read(r, binary.BigEndian, &originalLen); err != nil {
		return nil, fmt.Errorf("huffman: reading original length: %w", err)
	}

	if originalLen == 0 {
		return []byte{}, nil
	}

	var numSymbols uint16
	if err := binary.Read(r, binary.BigEndian, &numSymbols); err != nil {
		return nil, fmt.Errorf("huffman:reading symbol count: %w", err)
	}

	root := &Node{}

	for i := 0; i < int(numSymbols); i++ {
		sym, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("huffman: reading symbol #%d: %w", i, err)
		}

		codeLen, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("huffman: reading code length #%d: %w", i, err)
		}
		if codeLen == 0 {
			return nil, fmt.Errorf("huffman: symbol #%d has zero-length code", i)
		}

		nBytes := (int(codeLen) + 7) / 8
		codeBuf := make([]byte, nBytes)
		if _, err := io.ReadFull(r, codeBuf); err != nil {
			return nil, fmt.Errorf("huffman: reading code bits #%d: %w", i, err)
		}

		code := unpackCodeBits(codeBuf, int(codeLen))
		insertCode(root, sym, code)
	}

	br := NewBitReader(r)
	out := make([]byte, 0, originalLen)

	for uint32(len(out)) < originalLen {
		node := root
		for !node.IsLeaf() {
			bit, err := br.ReadBit()
			if err != nil {
				return nil, fmt.Errorf("huffman: reading bit at output offset %d: %w", len(out), err)
			}
			if bit == 0 {
				node = node.Left
			} else {
				node = node.Right
			}
			if node == nil {
				return nil, fmt.Errorf("huffman: invalid code at output offset %d", len(out))
			}
		}
		out = append(out, node.Symbol)
	}

	return out, nil
}

func unpackCodeBits(data []byte, codeLen int) string {
	code := make([]byte, codeLen)
	for i := 0; i < codeLen; i++ {
		if data[i/8]&(1<<uint(7-i%8)) != 0 {
			code[i] = '1'
		} else {
			code[i] = '0'
		}
	}
	return string(code)
}

func insertCode(root *Node, sym byte, code string) {
	node := root
	for i := 0; i < len(code); i++ {
		if code[i] == '0' {
			if node.Left == nil {
				node.Left = &Node{}
			}
			node = node.Left
		} else {
			if node.Right == nil {
				node.Right = &Node{}
			}
			node = node.Right
		}
	}
	node.Symbol = sym
}
