package ats

import (
	"unsafe"
)

func TickFX2Bytes(buf *TickFX) []byte {
	cnt := int(unsafe.Sizeof(TickFX{}))
	res := (*(*[1 << 31]byte)(unsafe.Pointer(buf)))[:cnt]
	return res
}

func Bytes2TickFX(buf []byte) *TickFX {
	if len(buf) < int(unsafe.Sizeof(TickFX{})) {
		return nil
	}
	res := (*TickFX)(unsafe.Pointer(&buf[0]))
	return res
}
