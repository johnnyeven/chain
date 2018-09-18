package util

import (
	"net"
	"strings"
	"strconv"
	"errors"
)

func DecodeCompactIPPortInfo(info string) (ip net.IP, port int, err error) {
	if len(info) != 6 {
		err = errors.New("compact info should be 6-length long")
		return
	}

	ip = net.IPv4(info[0], info[1], info[2], info[3])
	port = int((uint16(info[4]) << 8) | uint16(info[5]))
	return
}

func EncodeCompactIPPortInfo(ip net.IP, port int) (info string, err error) {
	if port > 65535 || port < 0 {
		err = errors.New(
			"port should be no greater than 65535 and no less than 0")
		return
	}

	p := Int2Bytes(uint64(port))
	if len(p) < 2 {
		p = append(p, p[0])
		p[0] = 0
	}

	ip = ip.To4()
	info = string(append(ip, p...))
	return
}

func GenerateAddress(ip string, port int) string {
	return strings.Join([]string{ip, strconv.Itoa(port)}, ":")
}
