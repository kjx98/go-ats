package ats

import (
	"math"
	"testing"
)

func round(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}

func TestInitSymbols(t *testing.T) {
	initSymbols()
	t.Log("initTemp:")
	for i := 0; i < len(initTemp); i++ {
		t.Log(&initTemp[i])
	}
}

func TestInitTicks(t *testing.T) {
	initSymbols()
	t.Log("initTicks")
	for _, ti := range initTicks {
		t.Log(ti)
	}
}

func TestSymbolsFunc(t *testing.T) {
	initSymbols()
	newSymbolInfo("sh600600")
	newSymbolInfo("cu1903")
	newSymbolInfo("ESZ8")
	newSymbolInfo("ESY8")
	newSymbolInfo("GOOG")
	newSymbolInfo("SPY")
	//newSymbolInfo("EURUSD")
	//newSymbolInfo("EURJPY")
	//newSymbolInfo("USDJPY")
	newSymbolInfo("BTCUSD")
	t.Log("symInfos:")
	for _, symP := range symInfos {
		t.Log(symP)
	}
	if si, err := GetSymbolInfo("ESY8"); err == nil {
		t.Errorf("ESY8 shouldn't exist, %v", si)
	}
	if si, err := GetSymbolInfo("ESZ8"); err != nil {
		t.Error("not found ESZ8", err)
	} else if np := si.PriceNormal(2810.2534); np != 2810.25 {
		t.Errorf("%s NormalPrice 2810.2534 to %f", si.Ticker, np)
	} else if vv := si.CalcVolume(422000, 2810.25); vv != 60.0 {
		t.Errorf("%s CalcVolume: %f", si.Ticker, vv)
	} else {
		t.Logf("%s digits/volDigits: %d/%d", si.Ticker, si.Digits(), si.VolumeDigits())
	}
	if si, err := GetSymbolInfo("BTCUSD"); err != nil {
		t.Error("not found BTCUSD", err)
	} else if np := si.PriceNormal(32810.2734) - 32810.27; math.Abs(np*1e8) > 0.01 {
		t.Errorf("%s NormalPrice 32810.2734 to 32810.27, diff(%f)", si.Ticker, np)
	} else if vv := si.CalcVolume(422000, 32810.27); vv != 12.8618 {
		t.Errorf("%s CalcVolume: %f", si.Ticker, vv)
	} else {
		t.Logf("%s digits/volDigits: %d/%d", si.Ticker, si.Digits(), si.VolumeDigits())
	}
	if si, err := GetSymbolInfo("GOOG"); err != nil {
		t.Error("not found GOOG", err)
	} else if np := si.PriceNormal(2810.1234); np != 2810.12 {
		t.Errorf("%s NormalPrice 2810.1234 to %f", si.Ticker, np)
	} else if vv := si.CalcVolume(422000, 2810.25); vv != 150.0 {
		t.Errorf("%s CalcVolume: %f", si.Ticker, vv)
	} else {
		t.Logf("%s digits/volDigits: %d/%d", si.Ticker, si.Digits(), si.VolumeDigits())
	}
	if si, err := GetSymbolInfo("SPY"); err != nil {
		t.Error("not found SPY", err)
	} else if np := si.PriceNormal(263.0123); np != 263.01 {
		t.Errorf("%s NormalPrice 263.0123 to %f", si.Ticker, np)
	} else if vv := si.CalcVolume(31825, 263.01); vv != 121.0 {
		t.Errorf("%s CalcVolume: %f", si.Ticker, vv)
	} else {
		t.Logf("%s digits/volDigits: %d/%d", si.Ticker, si.Digits(), si.VolumeDigits())
	}
	if si, err := GetSymbolInfo("cu1903"); err != nil {
		t.Error("not found cu1903", err)
	} else if np := si.PriceNormal(45320); np != 45300.0 {
		t.Errorf("%s NormalPrice 45320 to %f", si.Ticker, np)
	} else if vv := si.CalcVolume(453210, 45320); vv != 20 {
		t.Errorf("%s CalcVolume: %d", si.Ticker, int(vv))
	} else {
		t.Logf("%s digits/volDigits: %d/%d", si.Ticker, si.Digits(), si.VolumeDigits())
	}
}

func TestSymbolInfo_CalcRiskVolume(t *testing.T) {
	type args struct {
		amt       float64
		riskPrice float64
	}
	tests := []struct {
		name string
		sym  string
		args args
		want float64
	}{
		// TODO: Add test cases.
		{"testRiskVol1", "sh600600", args{10000, 12.51}, 700},
		{"testRiskVol2", "cu1903", args{10000, 12.5}, 160},
		{"testRiskVol3", "ESZ8", args{10000, 12.52}, 15},
		{"testRiskVol4", "SPY", args{10000, 12.55}, 796},
		{"testRiskVol5", "EURUSD", args{10000, 12.8}, 0.01},
	}
	initSymbols()
	newSymbolInfo("sh600600")
	newSymbolInfo("cu1903")
	newSymbolInfo("ESZ8")
	newSymbolInfo("SPY")
	newSymbolInfo("EURUSD")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := GetSymbolInfo(tt.sym)
			if err != nil {
				t.Errorf("GetSymbolInfo(%s) %v", tt.sym, err)
				return
			}
			if got := s.CalcRiskVolume(tt.args.amt, tt.args.riskPrice); got != tt.want {
				t.Errorf("SymbolInfo.CalcRiskVolume() for %s = %v, want %v", tt.sym, got, tt.want)
			}
		})
	}
}

func TestSymbolInfo_CalcProfit(t *testing.T) {
	type args struct {
		openP  float64
		closeP float64
		volume int32
	}
	tests := []struct {
		name string
		sym  string
		args args
		want float64
	}{
		// TODO: Add test cases.
		{"testCalcProfit1", "sh600600", args{12.51, 12.53, 100}, 2.0},
		{"testCalcProfit2", "sh600600", args{12.51, 12.55, -200}, -8.0},
		{"testCalcProfit3", "cu1903", args{53000, 54000, 2}, 10000.0},
		{"testCalcProfit4", "ESZ8", args{2700, 2750, 1}, 2500.0},
		{"testCalcProfit5", "SPY", args{2780, 2800, -200}, -4000.0},
		{"testCalcProfit6", "EURUSD", args{1.1120, 1.1130, 1000}, 1000.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := GetSymbolInfo(tt.sym)
			if err != nil {
				t.Errorf("GetSymbolInfo(%s) %v", tt.sym, err)
				return
			}
			if got := s.CalcProfit(tt.args.openP, tt.args.closeP, tt.args.volume); round(got) != round(tt.want) {
				t.Errorf("SymbolInfo.CalcProfit() = %v, want %v", got, tt.want)
			}
		})
	}
}
