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
	st1 := julian.FromUint32(20170401)
	en1 := julian.FromUint32(20171231)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"LoadBarFX-EUR", args{"EURUSD", Min1, st1, en1}, false},
	}
	if noDukasData {
		tests[0].wantErr = true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadBarFX(tt.args.pair, tt.args.period, tt.args.startD, tt.args.endD); (err != nil) != tt.wantErr {
				t.Errorf("LoadBarFX() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cc, ok := cacheMinBar["EURUSD"]; ok {
				t.Logf("MinBar start: %d, end: %d, length: %d\n", cc.startD.Uint32(), cc.endD.Uint32(), len(cc.res))
				var ccTimeMs DateTimeMs
				if cnt := len(cc.res); cnt > 0 {
					t.Log("First Rec:", cc.res[0])
					t.Log("Last Rec:", cc.res[cnt-1])
					ccTimeMs = cc.res[cnt-1].Time.DateTimeMs()
				}
				if res, err := getBars("EURUSD", Hour1, ccTimeMs); err != nil {
					t.Error("getBars Hour1", err)
				} else if cnt := len(res.Date); cnt > 0 {
					t.Log("total getBars Hour1:", cnt)
					t.Log("First Rec:", res.Date[0], res.Open[0], res.High[0], res.Low[0], res.Volume[0])
					t.Log("Last Rec:", res.Date[cnt-1], res.Open[cnt-1], res.High[cnt-1], res.Low[cnt-1], res.Volume[cnt-1])
				}
			}
		})
	}
}
