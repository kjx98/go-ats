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
	Time(i int) timeT64
	Open(i int) float64
	High(i int) float64
	Low(i int) float64
	Close(i int) float64
	Volume(i int) float64
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

func Dates(s TaSeries) (res []timeT64) {
	if s.Len() <= 0 {
		return
	}
	ll := s.Len()
	res = make([]timeT64, ll)
	for i := 0; i < ll; i++ {
		res[i] = s.Time(i)
	}
	return
}

func Opens(s TaSeries) (res []float64) {
	if s.Len() <= 0 {
		return
	}
	ll := s.Len()
	res = make([]float64, ll)
	for i := 0; i < ll; i++ {
		res[i] = s.Open(i)
	}
	return
}

func Highs(s TaSeries) (res []float64) {
	if s.Len() <= 0 {
		return
	}
	ll := s.Len()
	res = make([]float64, ll)
	for i := 0; i < ll; i++ {
		res[i] = s.High(i)
	}
	return
}

func Lows(s TaSeries) (res []float64) {
	if s.Len() <= 0 {
		return
	}
	ll := s.Len()
	res = make([]float64, ll)
	for i := 0; i < ll; i++ {
		res[i] = s.Low(i)
	}
	return
}

func Closes(s TaSeries) (res []float64) {
	if s.Len() <= 0 {
		return
	}
	ll := s.Len()
	res = make([]float64, ll)
	for i := 0; i < ll; i++ {
		res[i] = s.Close(i)
	}
	return
}

func Volumes(s TaSeries) (res []float64) {
	if s.Len() <= 0 {
		return
	}
	ll := s.Len()
	res = make([]float64, ll)
	for i := 0; i < ll; i++ {
		res[i] = s.Volume(i)
	}
	return
}
