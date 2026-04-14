package archive

import "bytes"

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
