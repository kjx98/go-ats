package ats

import (
	"testing"

	"github.com/kjx98/golib/julian"
)

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
