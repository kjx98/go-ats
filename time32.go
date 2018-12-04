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
