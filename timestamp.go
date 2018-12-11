package ats

import (
	"time"
)

type timeT32 uint32

// return int64 value of time seconds from 1970/1/1
func (timeV timeT32) Time64() int64 {
	return int64(timeV)
}

// returns the UTC Time corresponding to the given Unix time, sec
//   seconds and nsec nanoseconds since January 1, 1970 UTC.
func (timeV timeT32) Time() time.Time {
	return time.Unix(int64(timeV), 0).UTC()
}

func (timeV timeT32) String() string {
	return timeV.Time().Format("01-02 15:04:05")
}

type timeT64 int64

// returns the UTC Time corresponding to the given Unix time, sec
//   seconds and nsec nanoseconds since January 1, 1970 UTC.
func (timeV timeT64) Time() time.Time {
	return time.Unix(int64(timeV), 0).UTC()
}

func (timeV timeT64) String() string {
	return timeV.Time().Format("2006-01-02 15:04:05")
}

func FromTime(t time.Time) timeT64 {
	return timeT64(t.Unix())
}
