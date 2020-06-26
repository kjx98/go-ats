package ats

import (
	"time"
	"unsafe"
)

type timeT32 uint32

// return int64 value of time seconds from 1970/1/1
// 		should adjust after dooms/uint32 overflow
func (timeV timeT32) Unix() int64 {
	return int64(timeV)
}

// Doomsday of uint32
//	currently is 2106-02-07
func (timeV timeT32) Dooms() time.Time {
	return time.Unix(int64(1)<<32, 0).UTC()
}

// returns the UTC Time corresponding to the given Unix time, sec
//   seconds and nsec nanoseconds since January 1, 1970 UTC.
func (timeV timeT32) Time() time.Time {
	return time.Unix(int64(timeV), 0).UTC()
}

func (timeV timeT32) String() string {
	return timeV.Time().Format("01-02 15:04:05")
}

// convert timeT32 to timeT64
func (timeV timeT32) TimeT64() (res timeT64) {
	res[0] = uint32(timeV)
	return
}

// convert Time to timeT32
//		should adjust after dooms
func timeT32FromTime(t time.Time) timeT32 {
	return timeT32(t.Unix())
}

type timeT64 [2]uint32

// return int64 value of time seconds from 1970/1/1
func (timeV timeT64) Unix() int64 {
	ret := (*int64)(unsafe.Pointer(&timeV[0]))
	return *ret
}

// returns the UTC Time corresponding to the given Unix time, sec
//   seconds and nsec nanoseconds since January 1, 1970 UTC.
func (timeV timeT64) Time() time.Time {
	return time.Unix(timeV.Unix(), 0).UTC()
}

func (timeV timeT64) String() string {
	return timeV.Time().Format("2006-01-02 15:04:05")
}

// timeT64 add int64 seconds
func (timeV timeT64) Add(v int64) timeT64 {
	r1 := (*int64)(unsafe.Pointer(&timeV[0]))
	res := *r1 + v
	ret := (*timeT64)(unsafe.Pointer(&res))
	return *ret
}

// convert timeT64 to timeT32
func (timeV timeT64) TimeT32() timeT32 {
	return timeT32(timeV[0])
}

func timeT64FromTime(t time.Time) timeT64 {
	res := t.Unix()
	ret := (*timeT64)(unsafe.Pointer(&res))
	return *ret
}

func timeT64FromInt64(res int64) timeT64 {
	ret := (*timeT64)(unsafe.Pointer(&res))
	return *ret
}
