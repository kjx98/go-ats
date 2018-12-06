package ats

import (
	"github.com/kjx98/golib/julian"
	"time"
)

type Bar interface {
	Date() time.Time
	Open() int32
	High() int32
	Low() int32
	Close() int32
	Volume() int64
}

type BarExt interface {
	Bar
	Turnover() float32
	OpenInterest() int32
}

type dayBar struct {
	date     julian.JulianDay
	open     int32
	high     int32
	low      int32
	vClose   int32
	turnover float32
	volume   int64
}

func (b *dayBar) Date() time.Time {
	return b.date.GetUTC()
}

func (b *dayBar) Open() int32 {
	return b.open
}

func (b *dayBar) High() int32 {
	return b.high
}

func (b *dayBar) Low() int32 {
	return b.low
}

func (b *dayBar) Close() int32 {
	return b.vClose
}

func (b *dayBar) Volume() int64 {
	return b.volume
}
