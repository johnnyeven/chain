package global

import (
	"bytes"
	"encoding/binary"
	"github.com/sirupsen/logrus"
)

func IntToByte(value uint64) []byte {
	buff := &bytes.Buffer{}
	err := binary.Write(buff, binary.BigEndian, value)
	if err != nil {
		logrus.Panicf("IntToByte value: %v, err: %v", value, err)
	}

	return buff.Bytes()
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
