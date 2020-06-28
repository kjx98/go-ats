package ats

import (
	"time"

	"github.com/kjx98/golib/julian"
)

// DateTimeMs ... millisecond since 1920/1/1
type DateTimeMs int64

// no consider for overflow int64 DateTimeMs, about 4.9e5 years
// Convert DateTimeMs

// Time returns the UTC Time corresponding to the given DateTimeMs time, sec
//   seconds and nsec nanoseconds since January 1, 1970 UTC.
func (dtMs DateTimeMs) Time() time.Time {
	ns := int64(dtMs%1000) * 1e6
	sec := int64(dtMs / 1000)
	return time.Unix(sec, ns).UTC()
}

// Unix   return seconds from 1970/1/1 UTC
func (dtMs DateTimeMs) Unix() int64 {
	return int64(dtMs / 1000)
}

// DateTimeMs Add delta
func (dtMs DateTimeMs) Add(v int) DateTimeMs {
	return DateTimeMs(int64(dtMs) + int64(v))
}

func (dtMs DateTimeMs) String() string {
	tt := dtMs.Time()
	return tt.Format("06-01-02 15:04:05.000")
}

// Millisecond ... return milli second 0..999
func (dtMs DateTimeMs) Millisecond() int {
	return int(dtMs % 1000)
}

func (t timeT64) DateTimeMs() DateTimeMs {
	return DateTimeMs(t.Unix() * 1000)
}

func (t timeT32) DateTimeMs() DateTimeMs {
	return DateTimeMs(t.Unix() * 1000)
}

// TimeToDateTimeMs ... convert time.Time to DateTimeMs
func TimeToDateTimeMs(dt time.Time) DateTimeMs {
	sec := dt.Unix() * 1000
	ms := dt.Nanosecond() / 1e6
	return DateTimeMs(sec + int64(ms))
}

// JulianToDateTimeMs ... convert jilian Day to DateTimeMs
func JulianToDateTimeMs(jDN julian.JulianDay) DateTimeMs {
	sec := jDN.UTC().Unix() * 1000
	return DateTimeMs(sec)
}
