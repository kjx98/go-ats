package ats

import (
	"time"
)

type DateTimeMs int64

// no consider for overflow int64 DateTimeMs, about 4.9e5 years
// Convert DateTimeMs
// returns the UTC Time corresponding to the given DateTimeMs time, sec
//   seconds and nsec nanoseconds since January 1, 1970 UTC.
func (dtMs DateTimeMs) Time() time.Time {
	ns := int64(dtMs&0x3ff) * 1e6
	sec := int64(dtMs >> 10)
	return time.Unix(sec, ns)
}

// return seconds from 1970/1/1 UTC
func (dtMs DateTimeMs) Unix() int64 {
	return int64(dtMs >> 10)
}

func (dtMs DateTimeMs) Millisecond() int {
	return int(dtMs & 0x3ff)
}

func (t timeT64) DateTimeMs() DateTimeMs {
	return DateTimeMs(int64(t) << 10)
}

// convert time.Time to DateTimeMs
func ToDateTimeMs(dt time.Time) DateTimeMs {
	sec := dt.Unix() << 10
	ms := dt.Nanosecond() / 1e6
	//return DateTimeMs(sec | int64(ms&0x3ff))
	return DateTimeMs(sec | int64(ms))
}
