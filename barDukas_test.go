package ats

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjx98/golib/julian"
)

func Test_loadTickFX(t *testing.T) {
	type args struct {
		pair   string
		startD julian.JulianDay
	}
	startT := julian.FromUint32(20170521)
	st1 := julian.FromUint32(20170529)
	tests := []struct {
		name       string
		args       args
		wantResLen int
		wantErr    bool
	}{
		// TODO: Add test cases.
		{"testEUR1", args{"EURUSD", startT}, 431095, false},
		{"testEUR2", args{"EURUSD", st1}, 0, true},
	}
	if noDukasData {
		tests[0].wantResLen = 0
		tests[0].wantErr = true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := loadTickFX(tt.args.pair, tt.args.startD)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadTick() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(gotRes) > 0 {
				t.Log("First Rec:", gotRes[0])
				t.Log("Last Rec:", gotRes[len(gotRes)-1])
			}
			if !reflect.DeepEqual(len(gotRes), tt.wantResLen) {
				t.Errorf("loadTick() = %v, want %v", len(gotRes), tt.wantResLen)
			}
		})
	}
}

func Test_loadMinFX(t *testing.T) {
	type args struct {
		pair   string
		startD julian.JulianDay
	}
	startT := julian.FromUint32(20170521)
	st1 := julian.FromUint32(20170529)
	tests := []struct {
		name       string
		args       args
		wantResLen int
		wantErr    bool
	}{
		// TODO: Add test cases.
		{"testEUR1", args{"EURUSD", startT}, 7199, false},
		{"testEUR2", args{"EURUSD", st1}, 0, true},
	}
	if noDukasData {
		tests[0].wantResLen = 0
		tests[0].wantErr = true
	}
	recSize := int(unsafe.Sizeof(minDT{}))
	if recSize != 36 {
		t.Logf("minDT size=%d, want 36", recSize)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := loadMinFX(tt.args.pair, tt.args.startD)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadMin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(gotRes) > 0 {
				cnt := len(gotRes)
				t.Log("Total minDT", cnt)
				t.Log("First Rec:", gotRes[0])
				t.Log("Last Rec:", gotRes[cnt-1])
			}
			if !reflect.DeepEqual(len(gotRes), tt.wantResLen) {
				t.Errorf("loadMin() = %v, want %v", len(gotRes), tt.wantResLen)
			}
		})
	}
}

func TestLoadTickFX(t *testing.T) {
	type args struct {
		pair   string
		startD julian.JulianDay
		endD   julian.JulianDay
		cnt    int
	}
	st1 := julian.FromUint32(20170403)
	en1 := julian.FromUint32(20171231)
	tests := []struct {
		name       string
		args       args
		wantResLen int
		wantErr    bool
	}{
		// TODO: Add test cases.
		{"testLoadTick1", args{"EURUSD", st1, 0, 3000000}, 3113157, false},
		//{"testLoadTick2", args{"EURUSD", 0, en1, 1000000}, 1171851, false},
		{"testLoadTick2", args{"EURUSD", 0, en1, 1000000}, 1173573, false},
	}
	// first tick 03-05-05 03:00:06.561@t440s, count 1171851
	// while 03-05-05 01:00:00.495@t410, count 1173573
	if noDukasData {
		tests[0].wantResLen = 0
		tests[0].wantErr = true
		tests[1].wantResLen = 0
		tests[1].wantErr = true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := LoadTickFX(tt.args.pair, tt.args.startD, tt.args.endD, tt.args.cnt)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTickFX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(gotRes) > 0 {
				cnt := len(gotRes)
				t.Log("Total tickDT", cnt)
				t.Log("First Rec:", gotRes[0])
				t.Log("Last Rec:", gotRes[cnt-1])
			}
			if !reflect.DeepEqual(len(gotRes), tt.wantResLen) {
				t.Errorf("LoadTickFX() = %v, want %v", len(gotRes), tt.wantResLen)
			}
		})
	}
}

func TestLoadMinFX(t *testing.T) {
	type args struct {
		pair   string
		startD julian.JulianDay
		endD   julian.JulianDay
		cnt    int
	}
	st1 := julian.FromUint32(20160103)
	en1 := julian.FromUint32(20170603)
	tests := []struct {
		name       string
		args       args
		wantResLen int
		wantErr    bool
	}{
		// TODO: Add test cases.
		{"testLoadMin1", args{"EURUSD", st1, 0, 60000}, 64786, false},
		{"testLoadMin2", args{"EURUSD", 0, en1, 1000000}, 1004620, false},
		{"testLoadMin3", args{"EURUSD", st1, en1, 0}, 531999, false},
		{"testLoadMin4", args{"GBPUSD", st1, en1, 0}, 532320, false},
		{"testLoadMin5", args{"USDJPY", st1, en1, 0}, 531563, false},
		{"testLoadMin6", args{"XAUUSD", st1, en1, 0}, 502411, false},
	}
	if noDukasData {
		tests[0].wantResLen = 0
		tests[0].wantErr = true
		tests[1].wantResLen = 0
		tests[1].wantErr = true
		tests[2].wantResLen = 0
		tests[2].wantErr = true
		tests[3].wantResLen = 0
		tests[3].wantErr = true
		tests[4].wantResLen = 0
		tests[4].wantErr = true
		tests[5].wantResLen = 0
		tests[5].wantErr = true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := LoadMinFX(tt.args.pair, tt.args.startD, tt.args.endD, tt.args.cnt)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadMinFX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(gotRes) > 0 {
				cnt := len(gotRes)
				t.Log("Total minDT", cnt)
				t.Log("First Rec:", gotRes[0])
				t.Log("Last Rec:", gotRes[cnt-1])
			}
			if !reflect.DeepEqual(len(gotRes), tt.wantResLen) {
				t.Errorf("LoadMinFX() = %v, want %v", len(gotRes), tt.wantResLen)
			}
		})
	}
	t.Log(DukasCacheStatus())
}
