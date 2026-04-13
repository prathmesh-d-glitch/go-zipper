package utils

var table [256]uint32

func init() {
	const polynomial = 0xEDB88320
	for i := 0; i < 256; i++ {
		crc := uint32(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ polynomial
			} else {
				crc >>= 1
			}
		}
		table[i] = crc
	}
}

func Checksum(data []byte) uint32 {
	crc := ^uint32(0)
	for _, b := range data {
		crc = (crc >> 8) ^ table[(crc^uint32(b))&0xFF]
	}
	return ^crc
}

func Verify(data []byte, expected uint32) bool {
	return Checksum(data) == expected
}
