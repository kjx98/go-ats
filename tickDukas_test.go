package ats

import (
	"testing"

	"github.com/kjx98/golib/julian"
)

func TestGetCurrencies(t *testing.T) {
	fx := getCurrencies()
	if len(fx) == 0 {
		if noDukasData {
			t.Log("no tickData")
			return
		}
		t.Error("getCurrencies return null slice")
	}
	t.Log("Got Currencies: ", fx)
	fn := getDirMax(homePath + "/forex/DukasTick/EURUSD/2017/09/01")
	if fn != "23h_ticks.bi5" {
		t.Error("max should be 23h_ticks.bi5, not", fn)
	}
	dd, hh := checkLastTick("EURUSD")
	t.Log("Last Tick for EURUSD", dd, " hour:", hh)
}

func TestLoadBi5Tick(t *testing.T) {
	st := julian.NewJulianDay(2017, 10, 1)
	if res, err := loadBi5TickFX("EURUSD", st, 22); err != nil {
		if noDukasData {
			t.Log("no tickData")
			return
		}
		t.Error("Load bi5 failed", err)
	} else if len(res) == 0 {
		t.Error("EURUSD for 2017/10/1 22:00 empty")
	} else {
		cnt := len(res)
		t.Log("bi5 record count:", cnt)
		t.Log("First tick:", res[0])
		t.Log("Last tick:", res[cnt-1])
	}
}

func TestOpenTickFX(t *testing.T) {
	type args struct {
		pair   string
		startD uint32
		endD   uint32
		maxCnt int
	}
	tests := []struct {
		name    string
		args    args
		wantRes int
		wantErr bool
	}{
		// TODO: Add test cases.
		{"TestOpenTickFX1", args{"EURUSD", 20120101, 20170601, 0}, 137079433, false},
		{"TestOpenTickFX1", args{"EURUSD", 20160103, 20170601, 0}, 137079433, false},
		{"TestOpenTickFX1", args{"EURUSD", 20120101, 20161231, 0}, 137079433, false},
		{"TestOpenTickFX2", args{"EURUSD", 20090104, 20121231, 0}, 71778793, false},
		{"TestOpenTickFX2", args{"EURUSD", 20090104, 20170601, 0}, 185193213, false},
	}
	if noDukasData {
		t.Log("no tickData")
		return
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := julian.FromUint32(tt.args.startD)
			et := julian.FromUint32(tt.args.endD)
			gotRes, err := OpenTickFX(tt.args.pair, st, et, tt.args.maxCnt)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenTickFX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRes.Len() != tt.wantRes {
				t.Errorf("OpenTickFX() = %v, want %v", gotRes.Len(), tt.wantRes)
			}
		})
	}
}
