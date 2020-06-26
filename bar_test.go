package ats

import (
	"testing"
	"time"
)

func TestPeriodBaseTime(t *testing.T) {
	tt := time.Now().UTC().Unix()
	// Min1
	t1, _ := periodBaseTime(tt, Min1)
	if t1%int64(Min1) != 0 {
		t.Error("Min1 not multiple of 60:", t1)
	}
	if td := timeT64FromInt64(t1).Time(); td.Second() != 0 {
		t.Error("Min1 second not zero:", td)
	} else {
		t.Log("Min1 base time:", td)
	}
	// Min5
	t1, _ = periodBaseTime(tt, Min5)
	if t1%int64(Min5) != 0 {
		t.Error("Min5 not multiple of ", int(Min5), t1)
	}
	if td := timeT64FromInt64(t1).Time(); td.Second() != 0 || td.Minute()%5 != 0 {
		t.Error("Min5 not multipe of 5min:", td)
	} else {
		t.Log("Min5 base time:", td)
	}
	// Hour1
	t1, _ = periodBaseTime(tt, Hour1)
	if t1%int64(Hour1) != 0 {
		t.Error("Hour1 not multiple of ", int(Hour1), t1)
	}
	if td := timeT64FromInt64(t1).Time(); td.Second() != 0 || td.Minute() != 0 {
		t.Error("Hour1 not multipe:", td)
	} else {
		t.Log("Hour1 base time:", td)
	}
	// Hour4
	t1, _ = periodBaseTime(tt, Hour4)
	if t1%int64(Hour1) != 0 {
		t.Error("Hour4 not multiple of Hour1", int(Hour1), t1)
	}
	if td := timeT64FromInt64(t1).Time(); td.Second() != 0 || td.Minute() != 0 {
		t.Error("Hour4 not hourly multiple:", td)
	} else {
		t.Log("Hour4 base time:", td)
	}
	// Daily
	t1, _ = periodBaseTime(tt, Daily)
	if t1%int64(Daily) != 0 {
		t.Error("Daily not multiple of ", int(Daily), t1)
	}
	if td := timeT64FromInt64(t1).Time(); td.Second() != 0 || td.Minute() != 0 || td.Hour() != 0 {
		t.Error("Daily not multipe:", td)
	} else {
		t.Log("Daily base time:", td)
	}
	// Weekly
	t1, _ = periodBaseTime(tt, Weekly)
	if t1%int64(Daily) != 0 {
		t.Error("Weekly not multiple of daily", int(Daily), t1)
	}
	if td := timeT64FromInt64(t1).Time(); td.Second() != 0 || td.Minute() != 0 || td.Hour() != 0 || td.Weekday() != 0 {
		t.Error("Weekly not sunday:", td)
	} else {
		t.Log("Weekly base time:", td)
	}
	// Monthly
	t1, _ = periodBaseTime(tt, Monthly)
	if t1%int64(Daily) != 0 {
		t1, _ = periodBaseTime(tt, Monthly)
		t.Error("Monthly not multiple of daily", int(Daily), t1)
	}
	if td := timeT64FromInt64(t1).Time(); td.Second() != 0 || td.Minute() != 0 || td.Hour() != 0 || td.Day() != 1 {
		t.Error("Monthly not 1st of month:", td)
	} else {
		t.Log("Monthly base time:", td)
	}
}
