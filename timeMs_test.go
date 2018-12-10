package ats

import (
	"testing"
	"time"
)

func TestDateTimeMs(t *testing.T) {
	t1 := time.Now()
	ms1 := ToDateTimeMs(t1)
	if t1.Unix() != ms1.Unix() || t1.Nanosecond()/1e6 != ms1.Millisecond() {
		t.Error("DateTime diff", t1.Unix(), ms1.Unix())
	}
}

func BenchmarkDateTimeMs(b *testing.B) {
	t1 := time.Now()
	for i := 0; i < b.N; i++ {
		ms := ToDateTimeMs(t1)
		_ = ms.Time()
	}
}
