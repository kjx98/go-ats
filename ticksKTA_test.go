package ats

import (
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

func TestTickSize(t *testing.T) {
	if tickHdSize := unsafe.Sizeof(TickHead{}); tickHdSize != 16 {
		t.Error("TickHead size diff")
	}
}

func TestTickCodec(t *testing.T) {
	ticks := make([]Tick, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	BaseT := uint32(time.Now().Unix())
	BaseP := 1000 + (r.Int31() & 0xffff)
	for idx := range ticks {
		tp := &ticks[idx]
		if idx == 0 {
			tp.Last = BaseP
		} else {
			pDelta := r.Uint32() & 0xffff
			tp.Last = BaseP + int32(int16(pDelta))
		}
		tp.Volume = r.Uint32()
		tp.Time = timeT32(BaseT)
		BaseT += (r.Uint32() & 0xfff)
	}

	var tickHd = &TickHead{}
	var tBuf []byte
	if buf, err := tickHd.EncodeTick(ticks); err == nil {
		tBuf = buf
		t.Log("Encoded size:", len(buf), "tickHd:", tickHd)
	} else {
		t.Error(err)
		return
	}
	if newTicks, err := tickHd.DecodeTick(tBuf); err != nil {
		t.Error(err)
	} else {
		if len(newTicks) != len(ticks) {
			t.Error("Decode len diff", len(newTicks))
			return
		}
		for idx := range ticks {
			tp := &ticks[idx]
			np := &newTicks[idx]
			if tp.Time != np.Time {
				t.Error("Rec Time diff", idx, tp.Time, np.Time)
			} else if tp.Last != np.Last {
				t.Error("rec Last diff", idx, tp.Last, np.Last)
			} else if tp.Volume != np.Volume {
				t.Error("rec Volume diff", idx, tp.Volume, np.Volume)
			}
		}
	}
}

func TestTickExtCodec(t *testing.T) {
	ticks := make([]TickExt, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	BaseT := uint32(time.Now().Unix())
	BaseP := 1000 + (r.Int31() & 0xffff)
	for idx := range ticks {
		tp := &ticks[idx]
		if idx == 0 {
			tp.Bid = BaseP
			tp.BidDepth = tp.Bid - int32(r.Uint32()&0xff)
		} else {
			pDelta := r.Uint32() & 0xffff
			tp.Bid = BaseP + int32(int16(pDelta))
			if pDelta < 256 {
				tp.BidDepth = tp.Bid - int32(r.Uint32()&0xff)
			} else {
				tp.BidDepth = tp.Bid - int32(r.Uint32()&0xffff)
			}
		}
		pDelta := r.Uint32() & 0xffff
		tp.Ask = BaseP + int32(int16(pDelta))
		if pDelta < 256 {
			tp.AskDepth = tp.Ask + int32(r.Uint32()&0xff)
		} else {
			tp.AskDepth = tp.Ask + int32(r.Uint32()&0xffff)
		}
		tp.BidVol = r.Uint32()
		tp.BidsVol = r.Uint32()
		tp.AskVol = r.Uint32()
		tp.AsksVol = r.Uint32()
		tp.Time = timeT32(BaseT)
		BaseT += (r.Uint32() & 0xfff)
	}

	var tickHd = &TickHead{}
	var tBuf []byte
	if buf, err := tickHd.EncodeTickExt(ticks); err == nil {
		tBuf = buf
		t.Log("Encoded size:", len(buf), "tickHd:", tickHd)
	} else {
		t.Error(err)
		return
	}
	if newTicks, err := tickHd.DecodeTickExt(tBuf); err != nil {
		t.Error(err)
	} else {
		if len(newTicks) != len(ticks) {
			t.Error("Decode len diff", len(newTicks))
			return
		}
		for idx := range ticks {
			tp := &ticks[idx]
			np := &newTicks[idx]
			if tp.Time != np.Time {
				t.Error("Rec Time diff", idx, tp.Time, np.Time)
			} else if tp.Bid != np.Bid {
				t.Error("rec Bid diff", idx, tp.Bid, np.Bid)
			} else if tp.BidVol != np.BidVol {
				t.Error("rec Volume diff", idx, tp.BidVol, np.BidVol)
			} else if tp.Ask != np.Ask {
				t.Error("rec Ask diff", idx, tp.Ask, np.Ask)
			} else if tp.AskVol != np.AskVol {
				t.Error("rec Volume diff", idx, tp.AskVol, np.AskVol)
			}
		}
	}
}

func TestMinTACodec(t *testing.T) {
	ticks := make([]MinTA, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	BaseT := uint32(time.Now().Unix())
	BaseP := 1000 + (r.Int31() & 0xffff)
	for idx := range ticks {
		tp := &ticks[idx]
		if idx == 0 {
			tp.Open = BaseP
		} else {
			pDelta := r.Uint32() & 0xffff
			tp.Open = BaseP + int32(int16(pDelta))
		}
		tp.Close = BaseP + int32(int16(r.Uint32()&0xffff))
		{
			// build high
			hh := tp.Open
			ll := tp.Close
			if hh < ll {
				hh, ll = ll, hh
			}
			pDelta := r.Uint32() & 0xffff
			h1 := BaseP + int32(int16(pDelta))
			pDelta = r.Uint32() & 0xffff
			l1 := BaseP + int32(int16(pDelta))
			if h1 < l1 {
				h1, l1 = l1, h1
			}
			if hh > h1 {
				tp.High = hh
			} else {
				tp.High = h1
			}
			if ll < l1 {
				tp.Low = ll
			} else {
				tp.Low = l1
			}
		}
		tp.Volume = r.Uint32()
		{
			ff := r.Uint32() & 0xf
			tp.UpVol = (tp.Volume * ff) >> 4
			tp.DownVol = tp.Volume - tp.UpVol
		}
		tp.Time = timeT32(BaseT)
		BaseT += (r.Uint32() & 0xfff)
	}

	var tickHd = &TickHead{}
	var tBuf []byte
	if buf, err := tickHd.EncodeMinTA(ticks); err == nil {
		tBuf = buf
		t.Log("Encoded size:", len(buf), "tickHd:", tickHd)
	} else {
		t.Error(err)
		return
	}
	if newTicks, err := tickHd.DecodeMinTA(tBuf); err != nil {
		t.Error(err)
	} else {
		if len(newTicks) != len(ticks) {
			t.Error("Decode len diff", len(newTicks))
			return
		}
		for idx := range ticks {
			tp := &ticks[idx]
			np := &newTicks[idx]
			if tp.Time != np.Time {
				t.Error("Rec Time diff", idx, tp.Time, np.Time)
			} else if tp.Open != np.Open {
				t.Error("rec Open diff", idx, tp.Open, np.Open)
			} else if tp.High != np.High {
				t.Error("rec High diff", idx, tp.High, np.High)
			} else if tp.Low != np.Low {
				t.Error("rec Low diff", idx, tp.Low, np.Low)
			} else if tp.Close != np.Close {
				t.Error("rec Close diff", idx, tp.Close, np.Close)
			} else if tp.Volume != np.Volume {
				t.Error("rec Volume diff", idx, tp.Volume, np.Volume)
			}
		}
	}
}
