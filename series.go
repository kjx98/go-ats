package ats

// Series data with int index
// Len return length of Series
// DataAt return Data at index i
//	if i >= 0,  data at index
//	else  data at index reverse order, -1 for last data
type Series interface {
	Len() int
	DataAt(i int) float64
}

// TimeSeries time as index
type TimeSeries interface {
	Len() int
	DataAt(timeT64) float64
	Index(timeT64) int
}

// TaSeries TA OHLCV
type TaSeries interface {
	Len() int
	BarValue(i int) (Ti timeT64, Op, Hi, Lo, Cl float64, Vol float64)
}

func NewSlice(s Series) (res []float64) {
	if s.Len() <= 0 {
		return
	}
	ll := s.Len()
	res = make([]float64, ll)
	for i := 0; i < ll; i++ {
		res[i] = s.DataAt(i)
	}
	return
}
