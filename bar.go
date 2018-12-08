package ats

import (
	_ "github.com/kjx98/golib/julian"
)

type Bars struct {
	sym    string // symbol name
	period int    // time period in second
	Date   []timeT64
	Open   []float64
	High   []float64
	Low    []float64
	Close  []float64
	Volume []float64
}
