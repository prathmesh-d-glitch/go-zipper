package archive

import "bytes"

type readerEntry struct {
	header     FileHeader
	dataOffset uint64
}

type Reader struct {
	r       *bytes.Reader
	entries []readerEntry
}

var minArchiveSize = len(Magic) + 1 + 2 + 8
