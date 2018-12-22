package ats

import (
	"errors"

	"github.com/kjx98/golib/julian"
)

type Tick struct {
	Time   timeT32
	Last   int32
	Volume uint32
}

type TickFX struct {
	Time DateTimeMs
	Bid  int32
	Ask  int32
}

type TickExt struct {
	Time     timeT32
	Bid      int32
	BidVol   uint32
	Ask      int32
	AskVol   uint32
	BidDepth int32
	AskDepth int32
	BidsVol  uint32
	AsksVol  uint32
}

type MinTA struct {
	Time    timeT32
	Open    int32
	High    int32
	Low     int32
	Close   int32
	Volume  uint32
	UpVol   uint32
	DownVol uint32
}

type MinFX struct {
	Time      timeT64
	Open      int32
	High      int32
	Low       int32
	Close     int32
	Ticks     int32
	Volume    float32
	SpreadMin uint16
	SpreadMax uint16
}

type MinTAExt struct {
	Time     timeT32
	Avg      int32
	Turnover float32
	OpenInt  uint32
}

type DayTA struct {
	Date     julian.JulianDay
	Open     int32
	High     int32
	Low      int32
	Close    int32
	Turnover float32
	Volume   int64
}

type minBarFX struct {
	startT timeT64
	endT   timeT64
	fMulti float64
	fDiv   float64
	ta     []MinFX
}

func (mt *minBarFX) Len() int {
	return len(mt.ta)
}

func (mt *minBarFX) Time(i int) timeT64 {
	return mt.ta[i].Time
}

func (mt *minBarFX) Open(i int) float64 {
	return float64(mt.ta[i].Open) * mt.fDiv
}

func (mt *minBarFX) High(i int) float64 {
	return float64(mt.ta[i].High) * mt.fDiv
}

func (mt *minBarFX) Low(i int) float64 {
	return float64(mt.ta[i].Low) * mt.fDiv
}

func (mt *minBarFX) Close(i int) float64 {
	return float64(mt.ta[i].Close) * mt.fDiv
}

// FX no actual volume, using ticks instead
func (mt *minBarFX) Volume(i int) float64 {
	return float64(mt.ta[i].Ticks)
}

type minBarTA struct {
	//period Period
	//startT timeT64
	//endT   timeT64
	fMulti float64
	fDiv   float64
	ta     []MinTA
}

func (mt *minBarTA) Len() int {
	return len(mt.ta)
}

func (mt *minBarTA) Time(i int) timeT64 {
	return timeT64(mt.ta[i].Time)
}

func (mt *minBarTA) Open(i int) float64 {
	return float64(mt.ta[i].Open) * mt.fDiv
}

func (mt *minBarTA) High(i int) float64 {
	return float64(mt.ta[i].High) * mt.fDiv
}

func (mt *minBarTA) Low(i int) float64 {
	return float64(mt.ta[i].Low) * mt.fDiv
}

func (mt *minBarTA) Close(i int) float64 {
	return float64(mt.ta[i].Close) * mt.fDiv
}

func (mt *minBarTA) Volume(i int) float64 {
	return float64(mt.ta[i].Volume)
}

var errBarPeriod = errors.New("Invalid period for baseBar")

func LoadBarFX(pair string, period Period, startD, endD julian.JulianDay) (err error) {
	if period != Min1 {
		return errBarPeriod
	}
	var mBar = minBarFX{}
	var fKey SymbolKey
	if si, err := GetSymbolInfo(pair); err != nil {
		return err
	} else {
		fKey = SymbolKey(si.fKey)
		mBar.fMulti = digitMulti(si.PriceDigits)
		mBar.fDiv = digitDiv(si.PriceDigits)
	}
	mBar.ta, err = LoadMinFX(pair, startD, endD, 0)
	if err != nil {
		return
	}

	startT := timeT64(startD.UTC().Unix())
	// end daily, to nextday minus 1 second
	endD++
	endT := timeT64(endD.UTC().Unix())
	endT--
	var bars = Bars{symKey: fKey, period: period, startDt: startT, endDt: endT}
	bars.Date = Dates(&mBar)
	bars.Open = Opens(&mBar)
	bars.High = Highs(&mBar)
	bars.Low = Lows(&mBar)
	bars.Close = Closes(&mBar)
	bars.Volume = Volumes(&mBar)
	return bars.loadBars(pair, period, mBar.startT, mBar.endT)
}
