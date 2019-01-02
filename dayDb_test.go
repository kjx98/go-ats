package ats

import (
	"testing"

	"github.com/kjx98/golib/julian"
)

func TestGetChart(t *testing.T) {
	type args struct {
		sym    string
		startD julian.JulianDay
		endD   julian.JulianDay
	}
	st1 := julian.FromUint32(20050301)
	en1 := julian.FromUint32(20181231)
	tests := []struct {
		name    string
		args    args
		wantRes int
	}{
		// TODO: Add test cases.
		{"GetChartETF50.1", args{"sh510050", 0, en1}, 3373},
		{"GetChartETF50.2", args{"sh510050", st1, en1}, 3373},
		{"GetChartSSI.1", args{"sh000001", 0, en1}, 6856},
		{"GetChartSSI.2", args{"sh000001", st1, en1}, 6856},
		{"GetChartSZI.1", args{"sz399001", 0, en1}, 6801},
		{"GetChartSZ0001", args{"sz000001", 0, en1}, 6584},
	}
	/*
		if _, err := OpenDB(); err != nil {
			t.Log("No mysql, no Test GetChart", err)
			return
		}
	*/
	if len(symbolsMap) == 0 {
		t.Log("no mysql connection")
		return
	}
	initSymbols()
	newSymbolInfo("sh510050")
	newSymbolInfo("sh000001")
	newSymbolInfo("sz399001")
	newSymbolInfo("sz000001")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := GetChart(tt.args.sym, tt.args.startD, tt.args.endD); len(gotRes) != tt.wantRes {
				t.Errorf("GetChart() len = %v, want %v", len(gotRes), tt.wantRes)
			} else if len(gotRes) > 0 {
				cnt := len(gotRes)
				t.Log("Total DayTA", cnt)
				t.Log("First Rec:", gotRes[0])
				t.Log("Last Rec:", gotRes[cnt-1])
			}
		})
	}
}
