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

type cacheMinFXType struct {
	startD julian.JulianDay
	endD   julian.JulianDay
	res    []MinFX
}

type cacheMinTAType struct {
	startD julian.JulianDay
	endD   julian.JulianDay
	res    []MinTA
}

type cacheDayTAType struct {
	startD julian.JulianDay
	endD   julian.JulianDay
	res    []DayTA
}

type cacheTAer interface {
	Len() int
	BarValue(row int) (ti timeT64, o, h, l, c int32, vol int64)
}

func (fxm *cacheMinFXType) Len() int {
	return len(fxm.res)
}

func (fxm *cacheMinFXType) BarValue(r int) (ti timeT64, o, h, l, c int32, vol int64) {
	if r < 0 || r > len(fxm.res) {
		return
	}
	ti, o, h, l, c, vol = fxm.res[r].Time, fxm.res[r].Open, fxm.res[r].High,
		fxm.res[r].Low, fxm.res[r].Close, int64(fxm.res[r].Ticks)
	return
}

func (fxm *cacheMinTAType) Len() int {
	return len(fxm.res)
}

func (fxm *cacheMinTAType) BarValue(r int) (ti timeT64, o, h, l, c int32, vol int64) {
	if r < 0 || r > len(fxm.res) {
		return
	}
	ti, o, h, l, c, vol = fxm.res[r].Time.TimeT64(), fxm.res[r].Open,
		fxm.res[r].High, fxm.res[r].Low, fxm.res[r].Close,
		int64(fxm.res[r].Volume)
	return
}

func (fxm *cacheDayTAType) Len() int {
	return len(fxm.res)
}

func (fxm *cacheDayTAType) BarValue(r int) (ti timeT64, o, h, l, c int32, vol int64) {
	if r < 0 || r > len(fxm.res) {
		return
	}
	ti, o, h, l, c, vol = timeT64FromTime(fxm.res[r].Date.UTC()),
		fxm.res[r].Open, fxm.res[r].High, fxm.res[r].Low,
		fxm.res[r].Close, int64(fxm.res[r].Volume)
	return
}

var cacheMinFX = map[string]cacheMinFXType{}
var cacheMinTA = map[string]cacheMinTAType{}
var cacheDayTA = map[string]cacheDayTAType{}

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

func (mt *minBarFX) BarValue(i int) (Ti timeT64, Op, Hi, Lo, Cl float64, Vol float64) {
	Ti, Op, Hi, Lo, Cl, Vol = mt.ta[i].Time, float64(mt.ta[i].Open)*mt.fDiv,
		float64(mt.ta[i].High)*mt.fDiv, float64(mt.ta[i].Low)*mt.fDiv,
		float64(mt.ta[i].Close)*mt.fDiv, float64(mt.ta[i].Ticks)
	return
}

type minBarTA struct {
	//period Period
	startT timeT64
	endT   timeT64
	fMulti float64
	fDiv   float64
	ta     []MinTA
}

func (mt *minBarTA) Len() int {
	return len(mt.ta)
}

func (mt *minBarTA) BarValue(i int) (Ti timeT64, Op, Hi, Lo, Cl float64, Vol float64) {
	Ti, Op, Hi, Lo, Cl, Vol = mt.ta[i].Time.TimeT64(), float64(mt.ta[i].Open)*mt.fDiv,
		float64(mt.ta[i].High)*mt.fDiv, float64(mt.ta[i].Low)*mt.fDiv,
		float64(mt.ta[i].Close)*mt.fDiv, float64(mt.ta[i].Volume)
	return
}

type dayBarTA struct {
	//period Period
	startT timeT64
	endT   timeT64
	fMulti float64
	fDiv   float64
	ta     []DayTA
}

func (mt *dayBarTA) Len() int {
	return len(mt.ta)
}

func (mt *dayBarTA) BarValue(i int) (Ti timeT64, Op, Hi, Lo, Cl float64, Vol float64) {
	Ti, Op, Hi, Lo, Cl, Vol = timeT64FromTime(mt.ta[i].Date.UTC()),
		float64(mt.ta[i].Open)*mt.fDiv, float64(mt.ta[i].High)*mt.fDiv,
		float64(mt.ta[i].Low)*mt.fDiv,
		float64(mt.ta[i].Close)*mt.fDiv, float64(mt.ta[i].Volume)
	return
}

var (
	errBarPeriod  = errors.New("Invalid period for baseBar")
	errOutOfBound = errors.New("Out of Tick Bound")
)

// LoadBarFX .. load forex Bar data
func LoadBarFX(pair string, period Period, startD, endD julian.JulianDay) (err error) {
	if period != Min1 {
		return errBarPeriod
	}
	var mBar = minBarFX{}
	var fKey SymbolKey
	if si, err := GetSymbolInfo(pair); err == nil {
		fKey = SymbolKey(si.fKey)
		mBar.fMulti = digitMulti(si.PriceDigits)
		mBar.fDiv = digitDiv(si.PriceDigits)
	} else {
		return err
	}
	mBar.ta, err = LoadMinFX(pair, startD, endD, 0)
	if err != nil {
		return
	}

	startT := timeT64FromTime(startD.UTC())
	// end daily, to nextday minus 1 second
	endD++
	endT := timeT64FromInt64(endD.UTC().Unix() - 1)
	var bars = Bars{symKey: fKey, period: period, startDt: startT, endDt: endT}
	cnt := mBar.Len()
	bars.Date = make([]timeT64, cnt)
	bars.Open = make([]float64, cnt)
	bars.High = make([]float64, cnt)
	bars.Low = make([]float64, cnt)
	bars.Close = make([]float64, cnt)
	bars.Volume = make([]float64, cnt)
	for i := 0; i < cnt; i++ {
		bars.Date[i], bars.Open[i], bars.High[i], bars.Low[i], bars.Close[i],
			bars.Volume[i] = mBar.BarValue(i)
	}
	/*
		bars.Date = Dates(&mBar)
		bars.Open = Opens(&mBar)
		bars.High = Highs(&mBar)
		bars.Low = Lows(&mBar)
		bars.Close = Closes(&mBar)
		bars.Volume = Volumes(&mBar)
	*/
	return bars.loadBars(pair, period, mBar.startT, mBar.endT)
}

// LoadDayBar ... load Bar data for symbol
func LoadDayBar(symbol string, period Period, startD, endD julian.JulianDay) (err error) {
	if period != Daily {
		return errBarPeriod
	}
	var mBar = dayBarTA{}
	var fKey SymbolKey
	if si, err := GetSymbolInfo(symbol); err == nil {
		fKey = SymbolKey(si.fKey)
		mBar.fMulti = digitMulti(si.PriceDigits)
		mBar.fDiv = digitDiv(si.PriceDigits)
	} else {
		return err
	}
	mBar.ta = GetChart(symbol, startD, endD)
	if len(mBar.ta) == 0 {
		return
	}

	startT := timeT64FromTime(startD.UTC())
	// end daily, to nextday minus 1 second
	endD++
	endT := timeT64FromInt64(endD.UTC().Unix() - 1)
	var bars = Bars{symKey: fKey, period: period, startDt: startT, endDt: endT}
	cnt := mBar.Len()
	bars.Date = make([]timeT64, cnt)
	bars.Open = make([]float64, cnt)
	bars.High = make([]float64, cnt)
	bars.Low = make([]float64, cnt)
	bars.Close = make([]float64, cnt)
	bars.Volume = make([]float64, cnt)
	for i := 0; i < cnt; i++ {
		bars.Date[i], bars.Open[i], bars.High[i], bars.Low[i], bars.Close[i],
			bars.Volume[i] = mBar.BarValue(i)
	}
	/*
		bars.Date = Dates(&mBar)
		bars.Open = Opens(&mBar)
		bars.High = Highs(&mBar)
		bars.Low = Lows(&mBar)
		bars.Close = Closes(&mBar)
		bars.Volume = Volumes(&mBar)
	*/
	return bars.loadBars(symbol, period, mBar.startT, mBar.endT)
}
