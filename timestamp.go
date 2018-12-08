package ats

import (
	"time"
)

type timeT32 uint32

func (timeV timeT32) Time64() int64 {
	return int64(timeV)
}

func (timeV timeT32) Time() time.Time {
	return time.Unix(int64(timeV), 0)
}

func (timeV timeT32) String() string {
	return timeV.Time().Format("01-02 15:04:05")
}

type timeT64 int64

func (timeV timeT64) Time() time.Time {
	return time.Unix(int64(timeV), 0)
}

func (timeV timeT64) String() string {
	return timeV.Time().Format("2006-01-02 15:04:05")
}

func FromTime(t time.Time) timeT64 {
	return timeT64(t.Unix())
}
