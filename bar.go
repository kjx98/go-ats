package ats

import (
	"errors"
	"fmt"
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
	symKey  SymbolKey // fastKey of symbol name
	period  Period    // time period in second
	startDt timeT64
	endDt   timeT64
	Date    []timeT64
	Open    []float64
	High    []float64
	Low     []float64
	Close   []float64
	Volume  []float64
}

// only cache for non basePeriod Bars
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
var minBarsBase []*Bars

// cache for Daily Base period Bars
var dayBarsBase []*Bars
var cacheBars = map[int]*BarCache{}

func (per Period) String() string {
	switch per {
	case Min1:
		return "Min1"
	case Min3:
		return "Min3"
	case Min5:
		return "Min5"
	case Min15:
		return "Min15"
	case Min30:
		return "Min30"
	case Hour1:
		return "Hour1"
	case Hour2:
		return "Hour2"
	case Hour4:
		return "Hour4"
	case Hour8:
		return "Hour8"
	case Daily:
		return "Daily"
	case Weekly:
		return "Weekly"
	case Monthly:
		return "Monthly"
	}
	return "Inv Period"
}

func (b *Bars) String() string {
	si, err := b.symKey.SymbolInfo()
	if err != nil {
		return fmt.Sprintf("Bars with error symKey: %s", err)
	}
	return fmt.Sprintf("%s, period %s, start: %s, end: %s, Total: %d", si.Ticker, b.period.String(),
		b.startDt.String(), b.endDt.String(), len(b.Date))
}

func (b *Bars) RowString(idx int) string {
	if idx < 0 || idx >= len(b.Date) {
		return "OOB idx"
	}
	si, err := b.symKey.SymbolInfo()
	if err != nil {
		return fmt.Sprintf("Bars with error symKey: %s", err)
	}
	dig := si.Digits()
	if dig <= 0 {
		return fmt.Sprintf("%s %d/%d/%d/%d %d", b.Date[idx].String(), int(b.Open[idx]),
			int(b.High[idx]), int(b.Low[idx]), int(b.Close[idx]), int(b.Volume[idx]))
	}
	return fmt.Sprintf("%s %.*f/%.*f/%.*f/%.*f %d", b.Date[idx].String(), dig,
		b.Open[idx], dig, b.High[idx], dig, b.Low[idx], dig, b.Close[idx],
		int(b.Volume[idx]))
}

// using int for BarCacheHash
func getBarCacheHash(fKey int, period Period) int {
	return (fKey << 16) | (int(period) & 0xffff)
}

// Load BaseBars cache for ATS
//	Min1/Min5 for internal daily
//	Daily for daily/weekly/monthly
func (b *Bars) loadBars(sym string, period Period, startDt, endDt timeT64) error {
	si, err := GetSymbolInfo(sym)
	if err != nil || si.fKey != b.symKey {
		return invalidSymbol
	}
	if len(b.Date) == 0 {
		return nil
	}
	switch period {
	case Min1, Min5:
		if cnt := len(minBarsBase); cnt < nInstruments {
			nb := make([]*Bars, nInstruments)
			if cnt > 0 {
				copy(nb, minBarsBase)
			}
			minBarsBase = nb
		}
	case Daily:
		if cnt := len(dayBarsBase); cnt < nInstruments {
			nb := make([]*Bars, nInstruments)
			if cnt > 0 {
				copy(nb, dayBarsBase)
			}
			dayBarsBase = nb
		}
	default:
		return invalidPeriod
	}
	if startDt == 0 || startDt > b.Date[0] {
		startDt = b.Date[0]
	}
	if endDt == 0 {
		cnt := len(b.Date)
		endDt = timeT64(int64(b.Date[cnt-1]) + int64(b.period))
	}
	b.symKey = SymbolKey(si.fKey)
	b.period = period
	b.startDt = startDt
	b.endDt = endDt
	switch period {
	case Min1, Min5:
		minBarsBase[si.fKey-1] = b
	case Daily:
		dayBarsBase[si.fKey-1] = b
	}

	return nil
}

func (b *Bars) timeBars(curTime DateTimeMs) *Bars {
	cnt := len(b.Date)
	if cnt == 0 {
		return b
	}
	lastT, _ := periodBaseTime(curTime.Unix(), b.period)
	if int64(b.Date[cnt-1]) < lastT {
		return b
	}
	for cnt > 0 {
		if int64(b.Date[cnt-1]) < lastT {
			break
		}
		cnt--
	}
	var newBar = Bars{symKey: b.symKey}
	newBar.period = b.period
	newBar.startDt = b.startDt
	newBar.endDt = timeT64(lastT)
	j := 0
	//optimize for backtesting, limit lookback
	/*
		if cnt > 512 {
			// only lookback 512 Bars, maybe 1024 better
			j = cnt - 512
		}
	*/
	newBar.Date = b.Date[j:cnt]
	newBar.Open = b.Open[j:cnt]
	newBar.High = b.High[j:cnt]
	newBar.Low = b.Low[j:cnt]
	newBar.Close = b.Close[j:cnt]
	newBar.Volume = b.Volume[j:cnt]
	return &newBar
}

// Get Bars for symbol with period
func getBars(sym string, period Period, curTime DateTimeMs) (res *Bars, err error) {
	si, err := GetSymbolInfo(sym)
	if si.fKey <= 0 {
		err = invalidSymbol
		return
	}
	res, err = getBarsByKey(int(si.fKey), period, curTime)
	return
}

// Get Bars by fastKey of symbol with period
func getBarsByKey(fKey int, period Period, curTime DateTimeMs) (res *Bars, err error) {
	var basePeriod Period
	switch period {
	case Min1, Min3:
		basePeriod = Min1
	case Min5, Min15, Min30:
		fallthrough
	case Hour1, Hour2, Hour4, Hour8:
		// try Min1 first
		basePeriod = Min1
	case Daily, Weekly, Monthly:
		basePeriod = Daily
	default:
		err = invalidPeriod
		return
	}

	var baseBars *Bars
	switch basePeriod {
	case Min5, Min1:
		if fKey > len(minBarsBase) {
			err = noCacheBase
			return
		}
		baseBars = minBarsBase[fKey-1]
		if baseBars == nil || period < baseBars.period {
			err = noCacheBase
			return
		}
		basePeriod = baseBars.period
	case Daily:
		if fKey > len(dayBarsBase) {
			err = noCacheBase
			return
		}
		baseBars = dayBarsBase[fKey-1]
		if baseBars == nil {
			err = noCacheBase
			return
		}
	}

	if period != baseBars.period {
		bkey := getBarCacheHash(fKey, period)
		if cc, ok := cacheBars[bkey]; ok {
			res = &cc.Bars
			if res.endDt >= baseBars.endDt {
				cc.lastAccess = timeT64(time.Now().Unix())
				res = res.timeBars(curTime)
				return
			}
			// baseBars get updated, resample
			delete(cacheBars, bkey)
		}
		if res, err = baseBars.reSample(period); err != nil {
			return
		}
		cc := BarCache{}
		cc.Bars = *res
		cc.lastAccess = timeT64(time.Now().Unix())
		cc.loadTime = cc.lastAccess
		cc.basePeriod = basePeriod
		cacheBars[bkey] = &cc
	} else {
		res = baseBars
	}
	res = res.timeBars(curTime)
	return
}

func periodBaseTime(t int64, period Period) (res int64, mon time.Month) {
	switch period {
	case Min1, Min3, Min5, Min15, Min30:
		fallthrough
	case Hour1, Hour2, Hour4, Hour8:
		fallthrough
	case Daily:
		res = t - (t % int64(period))
	/*
		case Hour2, Hour4, Hour8:
			res = t - (t % int64(Hour1))
		case Weekly: // first workday for weekly start
			res = t - (t % int64(Daily))
		case Monthly: // first workday for monthly start
			res = t - (t % int64(Daily))
			_, mon, _ = timeT64(t).Time().Date()
	*/
	case Weekly:
		res = t - (t % int64(Daily))
		if wday := timeT64(res).Time().Weekday(); wday != 0 {
			res -= int64(Daily) * int64(wday)
		}
	case Monthly:
		y, mon, _ := timeT64(t).Time().Date()
		tt := time.Date(y, mon, 1, 0, 0, 0, 0, time.UTC)
		res = tt.Unix()
	}
	return
}

// resample Bars
func (b *Bars) reSample(newPeriod Period) (res *Bars, err error) {
	if newPeriod < b.period {
		err = invalidPeriod
		return
	}
	var vOpen, vHigh, vLow, vClose, volume float64
	var vDate int64
	switch newPeriod {
	case Min3, Min5, Min15, Min30:
		fallthrough
	case Hour1, Hour2, Hour4, Hour8:
		fallthrough
	case Daily:
		fallthrough
	case Weekly:
		cnt := len(b.Open)
		res = &Bars{symKey: b.symKey}
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
				vDate, _ = periodBaseTime(int64(b.Date[i]), newPeriod)
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
				vDate, mon = periodBaseTime(int64(b.Date[i]), newPeriod)
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
	res.startDt = b.startDt
	res.endDt = b.endDt
	return
}
