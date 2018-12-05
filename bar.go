package ats

import (
	"time"
	"github.com/kjx98/golib/julian"
)

type Bar interface {
	Date()	time.Time
	Open()	int32
	High()	int32
	Low()	int32
	Close()	int32
	Volume() int64
}

type BarExt interface {
	Bar
	Turnover() float32
	OpenInterest()	int32
}

type dayBar struct {
	vDate	julian.JulianDay
	vOpen	int32
	vHigh	int32
	vLow		int32
	vClose	int32
	vTurnover float32
	vVolume	int64
}

func (b *dayBar) Date() time.Time {
	return b.vDate.GetUTC()
}

func (b *dayBar) Open() int32 {
	return b.vOpen
}

func (b *dayBar) High() int32 {
	return b.vHigh
}

func (b *dayBar) Low() int32 {
	return b.vLow
}

func (b *dayBar) Close() int32 {
	return b.vClose
}

func (b *dayBar) Volume() int64 {
	return b.vVolume
}
