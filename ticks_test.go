package ats

import (
	"testing"
	//"unsafe"

	"github.com/kjx98/golib/julian"
)

func TestLoadBarFX(t *testing.T) {
	type args struct {
		pair   string
		period Period
		startD julian.JulianDay
		endD   julian.JulianDay
	}
	st1 := julian.FromUint32(20160103)
	en1 := julian.FromUint32(20170601)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"LoadBarFX-EUR", args{"EURUSD", Min1, 0, en1}, false},
		{"LoadBarFX-GBP", args{"GBPUSD", Min1, st1, en1}, false},
		{"LoadBarFX-JPY", args{"USDJPY", Min1, st1, en1}, false},
		{"LoadBarFX-XAU", args{"XAUUSD", Min1, st1, en1}, false},
	}
	if noDukasData {
		tests[0].wantErr = true
		tests[1].wantErr = true
		tests[2].wantErr = true
		tests[3].wantErr = true
	}
	//t.Logf("TickFX size: %d\n", unsafe.Sizeof(TickFX{}))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadBarFX(tt.args.pair, tt.args.period, tt.args.startD, tt.args.endD); (err != nil) != tt.wantErr {
				t.Errorf("LoadBarFX() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cc, ok := cacheMinFX[tt.args.pair]; ok {
				t.Logf("%s MinBar start: %d, end: %d, length: %d\n",
					tt.args.pair, cc.startD.Uint32(), cc.endD.Uint32(),
					len(cc.res))
				var ccTimeMs DateTimeMs
				if cnt := len(cc.res); cnt > 0 {
					t.Log("First Rec:", cc.res[0])
					t.Log("Last Rec:", cc.res[cnt-1])
					ccTimeMs = cc.res[cnt-1].Time.DateTimeMs()
				}
				if res, err := getBars(tt.args.pair, Hour1, ccTimeMs); err != nil {
					t.Error("getBars Hour1", err)
				} else if cnt := len(res.Date); cnt > 0 {
					t.Log(res)
					t.Log("First Rec:", res.RowString(0))
					t.Log("Last Rec:", res.RowString(cnt-1))
				}
			}
		})
	}
	t.Log(DukasCacheStatus())
}

func TestLoadDayBar(t *testing.T) {
	type args struct {
		symbol string
		period Period
		startD julian.JulianDay
		endD   julian.JulianDay
	}
	st1 := julian.FromUint32(20050301)
	en1 := julian.FromUint32(20181231)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"LoadBarETF50", args{"sh510050", Daily, st1, en1}, false},
		{"LoadBarSSI", args{"sh000001", Daily, st1, en1}, false},
		{"LoadBarSZSI", args{"sz399001", Daily, st1, en1}, false},
		{"LoadBarSZ0001", args{"sz000001", Daily, st1, en1}, false},
	}
	if len(symbolsMap) == 0 {
		tests[0].wantErr = true
		tests[1].wantErr = true
		tests[2].wantErr = true
		tests[3].wantErr = true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadDayBar(tt.args.symbol, tt.args.period, tt.args.startD, tt.args.endD); (err != nil) != tt.wantErr {
				t.Errorf("LoadDayBar() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cc, ok := cacheDayTA[tt.args.symbol]; ok {
				t.Logf("%s DailyBar start: %d, end: %d, length: %d\n", tt.args.symbol,
					cc.startD.Uint32(), cc.endD.Uint32(), len(cc.res))
				var ccTimeMs DateTimeMs
				if cnt := len(cc.res); cnt > 0 {
					t.Log("First Rec:", cc.res[0])
					t.Log("Last Rec:", cc.res[cnt-1])
					ccTimeMs = JulianToDateTimeMs(cc.res[cnt-1].Date)
				}
				if res, err := getBars(tt.args.symbol, Daily, ccTimeMs); err != nil {
					t.Error("getBars Daily", err)
				} else if cnt := len(res.Date); cnt > 0 {
					t.Log(res)
					t.Log("First Rec:", res.RowString(0))
					t.Log("Last Rec:", res.RowString(cnt-1))
				}
			}
		})
	}
	t.Log(DayDbCacheStatus())
}
