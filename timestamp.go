package ats

import (
	"time"
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

// convert Time to timeT32
//		should adjust after dooms
func timeT32FromTime(t time.Time) timeT32 {
	return timeT32(t.Unix())
}

type timeT64 int64

// return int64 value of time seconds from 1970/1/1
func (timeV timeT64) Unix() int64 {
	return int64(timeV)
}

// returns the UTC Time corresponding to the given Unix time, sec
//   seconds and nsec nanoseconds since January 1, 1970 UTC.
func (timeV timeT64) Time() time.Time {
	return time.Unix(int64(timeV), 0).UTC()
}

func (timeV timeT64) String() string {
	return timeV.Time().Format("2006-01-02 15:04:05")
}

func timeT64FromTime(t time.Time) timeT64 {
	return timeT64(t.Unix())
}
