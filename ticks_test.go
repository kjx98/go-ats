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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadBarFX(tt.args.pair, tt.args.period, tt.args.startD, tt.args.endD); (err != nil) != tt.wantErr {
				t.Errorf("LoadBarFX() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cc, ok := cacheMinBar["EURUSD"]; ok {
				t.Logf("MinBar start: %d, end: %d, length: %d\n", cc.startD.Uint32(), cc.endD.Uint32(), len(cc.res))
			}
		})
	}
}
