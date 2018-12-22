package ats

import (
	"testing"

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
	en1 := julian.FromUint32(20170603)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"LoadBarFX-EUR", args{"EURUSD", Min1, st1, en1}, false},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadBarFX(tt.args.pair, tt.args.period, tt.args.startD, tt.args.endD); (err != nil) != tt.wantErr {
				t.Errorf("LoadBarFX() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cc, ok := cacheMinDukas[tt.args.pair]; ok {
				t.Logf("%s MinBar start: %d, end: %d, length: %d\n", tt.args.pair,
					cc.startD.Uint32(), cc.endD.Uint32(), len(cc.res))
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
