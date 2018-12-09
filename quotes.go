package ats

// No Level2 quotes yet
type Quotes struct {
	TodayOpen float64
	TodayHigh float64
	TodayLow  float64
	Pclose    float64
	Last      float64
	Volume    int64
	Turnover  float64
	Bid       float64
	Ask       float64
	BidVol    int64
	AskVol    int64
}

// return quotes for symbol
func (s *SymbolInfo) GetQuotes() Quotes {
	return s.quote
}

// return ref for quotes of symbol, used by broker quotes feed
func getQuotesPtr(sym string) (qq *Quotes) {
	if si, err := GetSymbolInfo(sym); err == nil {
		qq = &si.quote
	}
	return
}

/*
func updateLastSales(sym string, last float64, vol float64) {
	if si, err := GetSymbolInfo(sym); err != nil {
		return
	} else {
		if si.VolDigits > 0 {
			vol *= digitDiv(si.VolDigits)
		}
		nVol := int64(vol)
		if nVol < si.quote.Volume {
			// volume must same or increased
			return
		}
		si.quote.Last = last
		si.quote.Volume = nVol
		if si.quote.TodayLow == 0 || last < si.quote.TodayLow {
			si.quote.TodayLow = last
		}
		if si.quote.TodayHigh < last {
			si.quote.TodayHigh = last
		}
	}
}

func updateBidAsk(sym string, bid, ask float64) {
	if si, err := GetSymbolInfo(sym); err != nil {
		return
	} else {
		si.quote.Bid = bid
		si.quote.Ask = ask
	}
}
*/
