package ats

import (
	"unsafe"
)

func TickFX2Bytes(buf []TickFX) []byte {
	cnt := len(buf) * int(unsafe.Sizeof(TickFX{}))
	res := (*(*[1 << 31]byte)(unsafe.Pointer(&buf[0])))[:cnt]
	return res
}

func Bytes2TickFX(buf []byte) []TickFX {
	cnt := len(buf) / int(unsafe.Sizeof(TickFX{}))
	res := (*(*[1 << 31]TickFX)(unsafe.Pointer(&buf[0])))[:cnt]
	return res
}
