package inet

import (
	"encoding/binary"
	"net"
	"unsafe"
)

func Ntohl(i uint32) uint32 {
	return binary.BigEndian.Uint32((*(*[4]byte)(unsafe.Pointer(&i)))[:])
}

func Htonl(i uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return *(*uint32)(unsafe.Pointer(&b[0]))
}

func Ntohs(i uint16) uint16 {
	return binary.BigEndian.Uint16((*(*[2]byte)(unsafe.Pointer(&i)))[:])
}

func Htons(i uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return *(*uint16)(unsafe.Pointer(&b[0]))
}

func Be64toh(i uint64) uint64 {
	return binary.BigEndian.Uint64((*(*[8]byte)(unsafe.Pointer(&i)))[:])
}

func Addr(s string) uint32 {
	return binary.BigEndian.Uint32(net.ParseIP(s).To4())
}

func Ntoa4(p unsafe.Pointer) string {
	return net.IP((*(*[net.IPv4len]byte)(p))[:]).String()
}

func Ntoa6(p unsafe.Pointer) string {
	return net.IP((*(*[net.IPv6len]byte)(p))[:]).String()
}
