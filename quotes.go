package ats

// Quotes ... No Level2 quotes yet
type Quotes struct {
	UpdateTime DateTimeMs
	TodayOpen  float64
	TodayHigh  float64
	TodayLow   float64
	Pclose     float64
	Last       float64
	Volume     int64
	Turnover   float64
	Bid        float64
	Ask        float64
	BidVol     int64
	AskVol     int64
}

// QuoteSubT ... subscribe quote struct
type QuoteSubT struct {
	Symbol    string
	QuotesPtr *Quotes
}

// QuoteEvent used by broker to notify quote/tick/bar update
// EventID    0   for quote/tick update, else bar period
type QuoteEvent struct {
	Symbol  string
	EventID int
}

// no export func for update quotes
// Feed should update quote using buffer pointed by QuoteSubType
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
