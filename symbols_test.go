package ats

import (
	"testing"
	"math"
)

func TestInitSymbols(t *testing.T) {
	initSymbols()
	t.Log("initTemp:")
	for i := 0; i < len(initTemp); i++ {
		t.Log(&initTemp[i])
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
	newSymbolInfo("EURUSD")
	newSymbolInfo("USDJPY")
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
	} else if np := si.PriceNormal(32810.2734)-32810.27; math.Abs(np*1e8) > 0.01 {
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
