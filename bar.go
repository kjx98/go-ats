package ats

import (
	"errors"
	"time"
)

// seconds for Bar time interval
type Period int32

const (
	// Min1 - 1 Minute time period
	Min1 Period = 60
	// Min3 - 3 Minute time period
	Min3 Period = 180
	// Min5 - 5 Minute time period
	Min5 Period = 300
	// Min15 - 15 Minute time period
	Min15 Period = 900
	// Min30 - 30 Minute time period
	Min30 Period = 1800
	// Hour1/Min60 - 60 Minute time period
	Hour1 Period = 3600
	// Hour2 - 2 hour time period
	Hour2 Period = 7200
	// Hour4 - 4 hour time period
	Hour4 Period = 14400
	// Hour8 - 8 hour time period
	Hour8 Period = 28800
	// Daily time period, using ymd/julianDay
	Daily Period = 86400
	// Weekly time period
	Weekly Period = 604800
	// Monthly time period, 30days, actual use ymd
	Monthly Period = 2592000
)

// Bars struct for talib
type Bars struct {
	symKey int32  // fastKey of symbol name
	period Period // time period in second
	Date   []timeT64
	Open   []float64
	High   []float64
	Low    []float64
	Close  []float64
	Volume []float64
}

// unimplemented yet
type BarCache struct {
	loadTime   timeT64
	lastAccess timeT64
	basePeriod Period
	Bars
}

var invalidPeriod = errors.New("Invalid Period")
var invalidSymbol = errors.New("Symbol not exist")
var noCacheBase = errors.New("No cache base Bars")

// cache for Min1/Min5 Base period Bars
var MinBarsBase []*Bars

// cache for Daily Base period Bars
var DayBarsBase []*Bars

// using int for BarCacheHash
func getBarCacheHash(fKey int, period Period) int {
	return (fKey << 16) | (int(period) & 0xffff)
}

// Load BaseBars cache for ATS
//	Min1/Min5 for internal daily
//	Daily for daily/weekly/monthly
func (b *Bars) loadBars(sym string, period Period) error {
	si, err := GetSymbolInfo(sym)
	if err != nil || si.fKey <= 0 {
		return invalidSymbol
	}
	switch period {
	case Min1:
		fallthrough
	case Min5:
		if cnt := len(MinBarsBase); cnt < nInstruments {
			nb := make([]*Bars, nInstruments)
			if cnt > 0 {
				copy(nb, MinBarsBase)
			}
			MinBarsBase = nb
		}
		b.symKey = int32(si.fKey)
		b.period = period
		MinBarsBase[si.fKey-1] = b
	case Daily:
		if cnt := len(DayBarsBase); cnt < nInstruments {
			nb := make([]*Bars, nInstruments)
			if cnt > 0 {
				copy(nb, DayBarsBase)
			}
			DayBarsBase = nb
		}
		b.symKey = int32(si.fKey)
		b.period = period
		DayBarsBase[si.fKey-1] = b
	default:
		return invalidPeriod
	}
	return nil
}

// Get Bars for symbol with period
func GetBars(sym string, period Period) (res *Bars, err error) {
	si, err := GetSymbolInfo(sym)
	if si.fKey <= 0 {
		return nil, invalidSymbol
	}
	res, err = GetBarsByKey(si.fKey, period)
	return
}

// Get Bars by fastKey of symbol with period
func GetBarsByKey(fKey int, period Period) (res *Bars, err error) {
	var basePeriod Period
	if err != nil {
		return
	}
	switch period {
	case Min1:
		fallthrough
	case Min3:
		basePeriod = Min1
	case Min5:
		fallthrough
	case Min15:
		fallthrough
	case Min30:
		fallthrough
	case Hour1:
		fallthrough
	case Hour2:
		fallthrough
	case Hour4:
		fallthrough
	case Hour8:
		// try Min1 first
		basePeriod = Min1
	case Daily:
		fallthrough
	case Weekly:
		fallthrough
	case Monthly:
		basePeriod = Daily
	default:
		return nil, invalidPeriod
	}

	var baseBars *Bars
	switch basePeriod {
	case Min5:
		fallthrough
	case Min1:
		if fKey > len(MinBarsBase) {
			return nil, noCacheBase
		}
		baseBars = MinBarsBase[fKey-1]
		if baseBars == nil || period < baseBars.period {
			return nil, noCacheBase
		}
		basePeriod = baseBars.period
	case Daily:
		if fKey > len(DayBarsBase) {
			return nil, noCacheBase
		}
		baseBars = DayBarsBase[fKey-1]
		if baseBars == nil {
			return nil, noCacheBase
		}
	}
	if period == baseBars.period {
		return baseBars, nil
	}
	return baseBars.reSample(period)
}

// resample Bars
func (b *Bars) reSample(newPeriod Period) (res *Bars, err error) {
	if newPeriod < b.period {
		return nil, invalidPeriod
	}
	var vOpen, vHigh, vLow, vClose, volume float64
	var vDate int64
	switch newPeriod {
	case Min3:
		fallthrough
	case Min5:
		fallthrough
	case Min15:
		fallthrough
	case Min30:
		fallthrough
	case Hour1:
		fallthrough
	case Hour2:
		fallthrough
	case Hour4:
		fallthrough
	case Hour8:
		fallthrough
	case Daily:
		fallthrough
	case Weekly:
		cnt := len(b.Open)
		res = &Bars{}
		res.period = newPeriod
		for i := 0; i < cnt; i++ {
			if int64(b.Date[i]) >= vDate+int64(newPeriod) {
				// new Bar
				if vDate != 0 {
					res.Date = append(res.Date, timeT64(vDate))
					res.Open = append(res.Open, vOpen)
					res.High = append(res.High, vHigh)
					res.Low = append(res.Low, vLow)
					res.Close = append(res.Close, vClose)
					res.Volume = append(res.Volume, volume)
				}
				vDate = 0
				vHigh = 0
				vLow = 0
				volume = 0
			}
			if vDate == 0 {
				vDate = int64(b.Date[i])
				vDate -= (vDate % int64(newPeriod))
				vOpen = b.Open[i]
			}
			if vHigh == 0 || b.High[i] > vHigh {
				vHigh = b.High[i]
			}
			if vLow == 0 || b.Low[i] < vLow {
				vLow = b.Low[i]
			}
			vClose = b.Close[i]
			volume += b.Volume[i]
		}
		if vDate != 0 {
			res.Date = append(res.Date, timeT64(vDate))
			res.Open = append(res.Open, vOpen)
			res.High = append(res.High, vHigh)
			res.Low = append(res.Low, vLow)
			res.Close = append(res.Close, vClose)
			res.Volume = append(res.Volume, volume)
		}
	case Monthly:
		cnt := len(b.Open)
		res = &Bars{}
		res.period = newPeriod
		var mon time.Month
		for i := 0; i < cnt; i++ {
			if mon != b.Date[i].Time().Month() {
				// new Bar
				if vDate != 0 {
					res.Date = append(res.Date, timeT64(vDate))
					res.Open = append(res.Open, vOpen)
					res.High = append(res.High, vHigh)
					res.Low = append(res.Low, vLow)
					res.Close = append(res.Close, vClose)
					res.Volume = append(res.Volume, volume)
				}
				vDate = 0
				vHigh = 0
				vLow = 0
				volume = 0
			}
			if vDate == 0 {
				y, m, _ := b.Date[i].Time().Date()
				tt := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
				mon = m
				vDate = tt.Unix()
				vOpen = b.Open[i]
			}
			if vHigh == 0 || b.High[i] > vHigh {
				vHigh = b.High[i]
			}
			if vLow == 0 || b.Low[i] < vLow {
				vLow = b.Low[i]
			}
			vClose = b.Close[i]
			volume += b.Volume[i]
		}
		if vDate != 0 {
			res.Date = append(res.Date, timeT64(vDate))
			res.Open = append(res.Open, vOpen)
			res.High = append(res.High, vHigh)
			res.Low = append(res.Low, vLow)
			res.Close = append(res.Close, vClose)
			res.Volume = append(res.Volume, volume)
		}
	}
	return res, nil
}
