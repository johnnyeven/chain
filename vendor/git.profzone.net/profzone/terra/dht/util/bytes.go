package util

import "crypto/rand"

func RandomString(size int) string {
	buff := make([]byte, size)
	rand.Read(buff)
	return string(buff)
}

func Int2Bytes(val uint64) []byte {
	data, j := make([]byte, 8), -1
	for i := 0; i < 8; i++ {
		shift := uint64((7 - i) * 8)
		data[i] = byte((val & (0xff << shift)) >> shift)

		if j == -1 && data[i] != 0 {
			j = i
		}
	}

	if j != -1 {
		return data[j:]
	}
	return data[:1]
}
