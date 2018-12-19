package ats

import (
	"errors"
	"unsafe"
)

type TickHead struct {
	Time    timeT32 // timeStamp Base, timeT32 for first tick
	BaseP   int32   // BasePrice, Last for first tick or pclose
	Count   int32   // Tick/TickExt count
	BufSize int32   // BufSize for encoded(compressed)
}

func getUint16(b []byte) (res uint16) {
	_ = b[1]
	bp := (*[2]byte)(unsafe.Pointer(&res))
	copy(bp[:2], b)
	return
}

func putUint16(b []byte, v uint16) {
	_ = b[1]
	bp := (*[2]byte)(unsafe.Pointer(&v))
	copy(b, bp[:2])
}

func getUint32(b []byte) (res uint32) {
	_ = b[3]
	bp := (*[4]byte)(unsafe.Pointer(&res))
	copy(bp[:4], b)
	return
}

func putUint32(b []byte, v uint32) {
	_ = b[3]
	bp := (*[4]byte)(unsafe.Pointer(&v))
	copy(b, bp[:4])
}

func updateTickHd(buf []byte, tickHd *TickHead) {
	hdBytes := (*[16]byte)(unsafe.Pointer(tickHd))
	copy(buf, hdBytes[:])
}

var errEmpty = errors.New("Empty Tick/MinTA buf")
var errNoWay = errors.New("Tick should no way go here")
var errTickCount = errors.New("tick Count diff")
var errBasePriceZero = errors.New("BasePrice is zero")
var errBaseTimeZero = errors.New("Base Time is zero")
var errCountZero = errors.New("TickHd.Count is zero")
var errBufSizeLen = errors.New("Head BufSize != bufLen")
var errTimeDelta = errors.New("Tick TimeDelta too large")
var errMinPrice = errors.New("MinTA price too large")
var errTickLast = errors.New("Tick lastDelta too large")
var errTickBid = errors.New("Tick BidDelta too large")
var errTickAsk = errors.New("Tick AskDelta too large")

// (* TickHead) EncodeTick([]Tick) ([]byte, error)
//	Encode/Compress Tick Slice to byte slice
func (tickHd *TickHead) EncodeTick(ticks []Tick) (buf []byte, err error) {
	nTicks := len(ticks)
	if nTicks == 0 {
		err = errEmpty
		return
	}
	tickHd.Time = ticks[0].Time
	tickHd.BaseP = ticks[0].Last
	tickHd.Count = int32(nTicks)
	if tickHd.BaseP == 0 {
		err = errBasePriceZero
		return
	}
	if tickHd.Time == 0 {
		err = errBaseTimeZero
		return
	}
	// encode Head
	buf = make([]byte, 16)
	updateTickHd(buf, tickHd)
	tmpBuf := make([]byte, 32)
	timeB := tickHd.Time
	for idx := range ticks {
		tickP := &ticks[idx]
		off := 1
		var fl byte
		timeDelta := tickP.Time - timeB
		/*
			if tickP.Time < timeB {
				timeDelta = 0
			}
		*/
		timeB = tickP.Time
		switch {
		case timeDelta < 256:
			//fl  = 0
			tmpBuf[off] = byte(timeDelta)
			off++
		case timeDelta < 65536:
			fl |= 0x80
			// copy 2 bytes
			putUint16(tmpBuf[off:], uint16(timeDelta))
			off += 2
		default:
			err = errTimeDelta

			return
		}
		lastDelta := tickP.Last - tickHd.BaseP
		switch {
		case lastDelta > -128 && lastDelta < 128:
			tmpBuf[off] = byte(lastDelta)
			off++
		case lastDelta > -32768 && lastDelta < 32768:
			fl |= 0x40
			putUint16(tmpBuf[off:], uint16(lastDelta))
			off += 2
		default:
			err = errTickLast
			return
		}
		switch {
		case tickP.Volume == 0:
			// do nothing
		case tickP.Volume < 256:
			fl |= 1
			tmpBuf[off] = byte(tickP.Volume)
			off++
		case tickP.Volume < 65536:
			fl |= 2
			putUint16(tmpBuf[off:], uint16(tickP.Volume))
			off += 2
		default:
			fl |= 3
			putUint32(tmpBuf[off:], tickP.Volume)
			off += 4
		}
		tmpBuf[0] = fl
		buf = append(buf, tmpBuf[:off]...)
	}
	tickHd.BufSize = int32(len(buf))
	updateTickHd(buf, tickHd)
	return
}

// (* TickHead) EncodeTickExt([]TickExt) ([]byte, error)
//	Encode/Compress TickExt Slice to byte slice
func (tickHd *TickHead) EncodeTickExt(ticks []TickExt) (buf []byte, err error) {
	nTicks := len(ticks)
	if nTicks == 0 {
		err = errEmpty
		return
	}
	tickHd.Time = ticks[0].Time
	tickHd.BaseP = ticks[0].Bid
	if tickHd.BaseP == 0 {
		tickHd.BaseP = ticks[0].Ask
	}
	tickHd.Count = int32(nTicks)
	if tickHd.BaseP == 0 {
		err = errBasePriceZero
		return
	}
	if tickHd.Time == 0 {
		err = errBaseTimeZero
		return
	}
	// encode Head
	buf = make([]byte, 16)
	updateTickHd(buf, tickHd)
	tmpBuf := make([]byte, 64)
	timeB := tickHd.Time
	for idx := range ticks {
		tickP := &ticks[idx]
		off := 1
		var fl byte
		timeDelta := tickP.Time - timeB
		/*
			if tickP.Time < timeB {
				timeDelta = 0
			}
		*/
		timeB = tickP.Time
		switch {
		case timeDelta < 256:
			//fl  = 0
			tmpBuf[off] = byte(timeDelta)
			off++
		case timeDelta < 65536:
			fl |= 0x80
			// copy 2 bytes
			putUint16(tmpBuf[off:], uint16(timeDelta))
			off += 2
		default:
			err = errTimeDelta
			return
		}
		lastDelta := tickP.Bid - tickHd.BaseP
		switch {
		case lastDelta > -128 && lastDelta < 128:
			tmpBuf[off] = byte(lastDelta)
			off++
			bDep := tickP.Bid - tickP.BidDepth
			if bDep > 255 {
				bDep = 255
			}
			tmpBuf[off] = byte(bDep)
			off++
		case lastDelta > -32768 && lastDelta < 32768:
			fl |= 0x40
			putUint16(tmpBuf[off:], uint16(lastDelta))
			off += 2
			putUint16(tmpBuf[off:], uint16(tickP.Bid-tickP.BidDepth))
			off += 2
		default:
			err = errTickBid
			return
		}
		lastDelta = tickP.Ask - tickHd.BaseP
		switch {
		case lastDelta > -128 && lastDelta < 128:
			tmpBuf[off] = byte(lastDelta)
			off++
			aDep := tickP.AskDepth - tickP.Ask
			if aDep > 255 {
				aDep = 255
			}
			tmpBuf[off] = byte(aDep)
			off++
		case lastDelta > -32768 && lastDelta < 32768:
			fl |= 0x20
			putUint16(tmpBuf[off:], uint16(lastDelta))
			off += 2
			putUint16(tmpBuf[off:], uint16(tickP.AskDepth-tickP.Ask))
			off += 2
		default:
			err = errTickAsk
			return
		}

		vol := tickP.BidVol
		if tickP.BidsVol > vol {
			vol = tickP.BidsVol
		}
		switch {
		case vol == 0:
			// do nothing
		case vol < 256:
			fl |= (1 << 2)
			tmpBuf[off] = byte(tickP.BidVol)
			off++
			tmpBuf[off] = byte(tickP.BidsVol)
			off++
		case vol < 65536:
			fl |= (2 << 2)
			putUint16(tmpBuf[off:], uint16(tickP.BidVol))
			off += 2
			putUint16(tmpBuf[off:], uint16(tickP.BidsVol))
			off += 2
		default:
			fl |= (3 << 2)
			putUint32(tmpBuf[off:], tickP.BidVol)
			off += 4
			putUint32(tmpBuf[off:], tickP.BidsVol)
			off += 4
		}
		vol = tickP.AskVol
		if tickP.AsksVol > vol {
			vol = tickP.AsksVol
		}
		switch {
		case vol == 0:
			// do nothing
		case vol < 256:
			fl |= 1
			tmpBuf[off] = byte(tickP.AskVol)
			off++
			tmpBuf[off] = byte(tickP.AsksVol)
			off++
		case vol < 65536:
			fl |= 2
			putUint16(tmpBuf[off:], uint16(tickP.AskVol))
			off += 2
			putUint16(tmpBuf[off:], uint16(tickP.AsksVol))
			off += 2
		default:
			fl |= 3
			putUint32(tmpBuf[off:], tickP.AskVol)
			off += 4
			putUint32(tmpBuf[off:], tickP.AsksVol)
			off += 4
		}
		tmpBuf[0] = fl
		buf = append(buf, tmpBuf[:off]...)
	}
	tickHd.BufSize = int32(len(buf))
	updateTickHd(buf, tickHd)
	return
}

// (* TickHead) EncodeMinTA([]MinTA) ([]byte, error)
//	Encode/Compress MinTA Slice to byte slice
func (tickHd *TickHead) EncodeMinTA(mins []MinTA) (buf []byte, err error) {
	nTicks := len(mins)
	if nTicks == 0 {
		err = errEmpty
		return
	}
	tickHd.Time = mins[0].Time
	tickHd.BaseP = mins[0].Open
	tickHd.Count = int32(nTicks)
	if tickHd.BaseP == 0 {
		err = errBasePriceZero
		return
	}
	if tickHd.Time == 0 {
		err = errBaseTimeZero
		return
	}
	// encode Head
	buf = make([]byte, 16)
	updateTickHd(buf, tickHd)
	tmpBuf := make([]byte, 32)
	timeB := tickHd.Time
	for idx := range mins {
		tickP := &mins[idx]
		off := 1
		var fl byte
		timeDelta := tickP.Time - timeB
		/*
			if tickP.Time < timeB {
				timeDelta = 0
			}
		*/
		timeB = tickP.Time
		switch {
		case timeDelta < 256:
			//fl  = 0
			tmpBuf[off] = byte(timeDelta)
			off++
		case timeDelta < 65536:
			fl |= 0x80
			// copy 2 bytes
			putUint16(tmpBuf[off:], uint16(timeDelta))
			off += 2
		default:
			err = errTimeDelta
			return
		}
		lastDelta := tickP.High - tickHd.BaseP
		if lastDelta <= 0 {
			lastDelta = tickP.Low - tickHd.BaseP
		} else {
			ll := tickP.Low - tickHd.BaseP
			if -ll > lastDelta {
				lastDelta = ll
			}
		}
		switch {
		case lastDelta > -128 && lastDelta < 128:
			{
				ll := tickP.Open - tickHd.BaseP
				tmpBuf[off] = byte(ll)
				off++
				ll = tickP.High - tickHd.BaseP
				tmpBuf[off] = byte(ll)
				off++
				ll = tickP.Low - tickHd.BaseP
				tmpBuf[off] = byte(ll)
				off++
				ll = tickP.Close - tickHd.BaseP
				tmpBuf[off] = byte(ll)
				off++
			}
		case lastDelta > -32768 && lastDelta < 32768:
			fl |= 0x40
			{
				ll := tickP.Open - tickHd.BaseP
				putUint16(tmpBuf[off:], uint16(ll))
				off += 2
				ll = tickP.High - tickHd.BaseP
				putUint16(tmpBuf[off:], uint16(ll))
				off += 2
				ll = tickP.Low - tickHd.BaseP
				putUint16(tmpBuf[off:], uint16(ll))
				off += 2
				ll = tickP.Close - tickHd.BaseP
				putUint16(tmpBuf[off:], uint16(ll))
				off += 2
			}
		default:
			err = errMinPrice
			return
		}
		switch {
		case tickP.Volume == 0:
			// do nothing
		case tickP.Volume < 256:
			fl |= 1
			tmpBuf[off] = byte(tickP.Volume)
			off++
			tmpBuf[off] = byte(tickP.UpVol)
			off++
			tmpBuf[off] = byte(tickP.DownVol)
			off++
		case tickP.Volume < 65536:
			fl |= 2
			putUint16(tmpBuf[off:], uint16(tickP.Volume))
			off += 2
			putUint16(tmpBuf[off:], uint16(tickP.UpVol))
			off += 2
			putUint16(tmpBuf[off:], uint16(tickP.DownVol))
			off += 2
		default:
			fl |= 3
			putUint32(tmpBuf[off:], tickP.Volume)
			off += 4
			fl |= 3
			putUint32(tmpBuf[off:], tickP.UpVol)
			off += 4
			fl |= 3
			putUint32(tmpBuf[off:], tickP.DownVol)
			off += 4
		}
		tmpBuf[0] = fl
		buf = append(buf, tmpBuf[:off]...)
	}
	tickHd.BufSize = int32(len(buf))
	updateTickHd(buf, tickHd)
	return
}

func ckBufLen(off, cnt, bLen int, ss string) (err error) {
	if off+cnt >= bLen {
		err = errors.New(ss)
	}
	return
}

// (* TickHead) DecodeTick([]Byte) ([]Tick, error)
//	Decode/DeCompress Tick Slice to byte slice
func (tickHd *TickHead) DecodeTick(buf []byte) (ticks []Tick, err error) {
	bLen := len(buf)
	if bLen == 0 {
		err = errEmpty
		return
	}
	if bLen != int(tickHd.BufSize) {
		err = errBufSizeLen
		return
	}
	hdSize := int(unsafe.Sizeof(TickHead{}))
	if tickHd.BaseP == 0 {
		err = errBasePriceZero
		return
	}
	if tickHd.Time == 0 {
		err = errBaseTimeZero
		return
	}
	nTicks := int(tickHd.Count)
	if nTicks == 0 {
		err = errCountZero
		return
	}
	ticks = make([]Tick, nTicks)
	off := hdSize
	idx := 0
	timeB := uint32(tickHd.Time)
	for off < bLen && idx < nTicks {
		fl := buf[off]
		if err = ckBufLen(off, 1, bLen, "Empty after flag"); err != nil {
			return
		}
		off++
		tp := &ticks[idx]
		idx++
		var timeDelta uint32
		errS := "Empty after rec.Time"
		if (fl & 0x80) != 0 {
			if err = ckBufLen(off, 2, bLen, errS); err != nil {
				return
			}
			timeDelta = uint32(getUint16(buf[off:]))
			off += 2
		} else {
			if err = ckBufLen(off, 1, bLen, errS); err != nil {
				return
			}
			timeDelta = uint32(buf[off])
			off++
		}
		timeB += timeDelta
		tp.Time = timeT32(timeB)
		//tp.Time = timeT32(timeB + timeDelta)
		//timeB = uint32(tp.Time)
		var lastDelta int32
		errS = "Empty after rec.Last"
		if (fl & 0x40) == 0 {
			if err = ckBufLen(off, 1, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int8(buf[off]))
			off++
		} else {
			if err = ckBufLen(off, 2, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int16(getUint16(buf[off:])))
			off += 2
		}
		tp.Last = tickHd.BaseP + lastDelta
		if (fl & 3) == 0 {
			continue
		}
		errS = "No enough bytes for rec.Volume"
		switch fl & 3 {
		case 1:
			if err = ckBufLen(off, 0, bLen, errS); err != nil {
				return
			}
			tp.Volume = uint32(buf[off])
			off++
		case 2:
			if err = ckBufLen(off, 1, bLen, errS); err != nil {
				return
			}
			tp.Volume = uint32(getUint16(buf[off:]))
			off += 2
		case 3:
			if err = ckBufLen(off, 3, bLen, errS); err != nil {
				return
			}
			tp.Volume = getUint32(buf[off:])
			off += 4
		}
	}
	if idx != int(tickHd.Count) {
		err = errTickCount

		return
	}
	if off != bLen {
		err = errNoWay

	}
	return
}

// (* TickHead) DecodeTickExt([]Byte) ([]TickExt, error)
//	Decode/DeCompress TickExt Slice to byte slice
func (tickHd *TickHead) DecodeTickExt(buf []byte) (ticks []TickExt, err error) {
	bLen := len(buf)
	if bLen == 0 {
		err = errEmpty
		return
	}
	if bLen != int(tickHd.BufSize) {
		err = errBufSizeLen
		return
	}
	hdSize := int(unsafe.Sizeof(TickHead{}))
	if tickHd.BaseP == 0 {
		err = errBasePriceZero
		return
	}
	if tickHd.Time == 0 {
		err = errBaseTimeZero
		return
	}
	nTicks := int(tickHd.Count)
	if nTicks == 0 {
		err = errCountZero
		return
	}
	ticks = make([]TickExt, nTicks)
	off := hdSize
	idx := 0
	timeB := uint32(tickHd.Time)
	for off < bLen && idx < nTicks {
		fl := buf[off]
		if err = ckBufLen(off, 1, bLen, "Empty after flag"); err != nil {
			return
		}
		off++
		tp := &ticks[idx]
		idx++
		var timeDelta uint32
		errS := "Empty after rec.Time"
		if (fl & 0x80) != 0 {
			if err = ckBufLen(off, 2, bLen, errS); err != nil {
				return
			}
			timeDelta = uint32(getUint16(buf[off:]))
			off += 2
		} else {
			if err = ckBufLen(off, 1, bLen, errS); err != nil {
				return
			}
			timeDelta = uint32(buf[off])
			off++
		}
		timeB += timeDelta
		tp.Time = timeT32(timeB)
		//tp.Time = timeT32(timeB + timeDelta)
		//timeB = uint32(tp.Time)
		var lastDelta int32
		var pDep int32
		errS = "Empty after rec.Bid"
		if (fl & 0x40) == 0 {
			if err = ckBufLen(off, 2, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int8(buf[off]))
			off++
			pDep = int32(buf[off])
			off++
		} else {
			if err = ckBufLen(off, 4, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int16(getUint16(buf[off:])))
			off += 2
			pDep = int32(getUint16(buf[off:]))
			off += 2
		}
		tp.Bid = tickHd.BaseP + lastDelta
		tp.BidDepth = tp.Bid - pDep
		errS = "Empty after rec.Ask"
		if (fl & 0x20) == 0 {
			if err = ckBufLen(off, 2, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int8(buf[off]))
			off++
			pDep = int32(buf[off])
			off++
		} else {
			if err = ckBufLen(off, 4, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int16(getUint16(buf[off:])))
			off += 2
			pDep = int32(getUint16(buf[off:]))
			off += 2
		}
		tp.Ask = tickHd.BaseP + lastDelta
		tp.AskDepth = tp.Ask + pDep
		errS = "Empty after rec.BidVol"
		switch (fl >> 2) & 3 {
		case 0:
			// do nothin
		case 1:
			if err = ckBufLen(off, 1, bLen, errS); err != nil {
				return
			}
			tp.BidVol = uint32(buf[off])
			off++
			tp.BidsVol = uint32(buf[off])
			off++
		case 2:
			if err = ckBufLen(off, 3, bLen, errS); err != nil {
				return
			}
			tp.BidVol = uint32(getUint16(buf[off:]))
			off += 2
			tp.BidsVol = uint32(getUint16(buf[off:]))
			off += 2
		case 3:
			if err = ckBufLen(off, 7, bLen, errS); err != nil {
				return
			}
			tp.BidVol = getUint32(buf[off:])
			off += 4
			tp.BidsVol = getUint32(buf[off:])
			off += 4
		}
		if (fl & 3) == 0 {
			continue
		}
		errS = "Empty after rec.AskVol"
		switch fl & 3 {
		case 0:
			// do nothin
		case 1:
			if err = ckBufLen(off, 1, bLen, errS); err != nil {
				return
			}
			tp.AskVol = uint32(buf[off])
			off++
			tp.AsksVol = uint32(buf[off])
			off++
		case 2:
			if err = ckBufLen(off, 3, bLen, errS); err != nil {
				return
			}
			tp.AskVol = uint32(getUint16(buf[off:]))
			off += 2
			tp.AsksVol = uint32(getUint16(buf[off:]))
			off += 2
		case 3:
			if err = ckBufLen(off, 7, bLen, errS); err != nil {
				return
			}
			tp.AskVol = getUint32(buf[off:])
			off += 4
			tp.AsksVol = getUint32(buf[off:])
			off += 4
		}
	}
	if idx != int(tickHd.Count) {
		err = errTickCount
		return
	}
	if off != bLen {
		err = errNoWay
	}
	return
}

// (* TickHead) DecodeMinTA([]Byte) ([]MinTA, error)
//	Decode/DeCompress MinTA Slice to byte slice
func (tickHd *TickHead) DecodeMinTA(buf []byte) (ticks []MinTA, err error) {
	bLen := len(buf)
	if bLen == 0 {
		err = errEmpty
		return
	}
	if bLen != int(tickHd.BufSize) {
		err = errBufSizeLen
		return
	}
	hdSize := int(unsafe.Sizeof(TickHead{}))
	if tickHd.BaseP == 0 {
		err = errBasePriceZero
		return
	}
	if tickHd.Time == 0 {
		err = errBaseTimeZero
		return
	}
	nTicks := int(tickHd.Count)
	if nTicks == 0 {
		err = errCountZero
		return
	}
	ticks = make([]MinTA, nTicks)
	off := hdSize
	idx := 0
	timeB := uint32(tickHd.Time)
	for off < bLen && idx < nTicks {
		fl := buf[off]
		if err = ckBufLen(off, 1, bLen, "Empty after flag"); err != nil {
			return
		}
		off++
		tp := &ticks[idx]
		idx++
		var timeDelta uint32
		errS := "Empty after rec.Time"
		if (fl & 0x80) != 0 {
			if err = ckBufLen(off, 2, bLen, errS); err != nil {
				return
			}
			timeDelta = uint32(getUint16(buf[off:]))
			off += 2
		} else {
			if err = ckBufLen(off, 1, bLen, errS); err != nil {
				return
			}
			timeDelta = uint32(buf[off])
			off++
		}
		timeB += timeDelta
		tp.Time = timeT32(timeB)
		//tp.Time = timeT32(timeB + timeDelta)
		//timeB = uint32(tp.Time)
		var lastDelta int32
		errS = "Empty after rec.OHLC"
		if (fl & 0x40) == 0 {
			if err = ckBufLen(off, 4, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int8(buf[off]))
			off++
			tp.Open = tickHd.BaseP + lastDelta
			lastDelta = int32(int8(buf[off]))
			off++
			tp.High = tickHd.BaseP + lastDelta
			lastDelta = int32(int8(buf[off]))
			off++
			tp.Low = tickHd.BaseP + lastDelta
			lastDelta = int32(int8(buf[off]))
			off++
			tp.Close = tickHd.BaseP + lastDelta
		} else {
			if err = ckBufLen(off, 8, bLen, errS); err != nil {
				return
			}
			lastDelta = int32(int16(getUint16(buf[off:])))
			off += 2
			tp.Open = tickHd.BaseP + lastDelta
			lastDelta = int32(int16(getUint16(buf[off:])))
			off += 2
			tp.High = tickHd.BaseP + lastDelta
			lastDelta = int32(int16(getUint16(buf[off:])))
			off += 2
			tp.Low = tickHd.BaseP + lastDelta
			lastDelta = int32(int16(getUint16(buf[off:])))
			off += 2
			tp.Close = tickHd.BaseP + lastDelta

		}
		errS = "No enough bytes for rec.Volume"
		switch fl & 3 {
		case 1:
			if err = ckBufLen(off, 2, bLen, errS); err != nil {
				return
			}
			tp.Volume = uint32(buf[off])
			off++
			tp.UpVol = uint32(buf[off])
			off++
			tp.DownVol = uint32(buf[off])
			off++
		case 2:
			if err = ckBufLen(off, 5, bLen, errS); err != nil {
				return
			}
			tp.Volume = uint32(getUint16(buf[off:]))
			off += 2
			tp.UpVol = uint32(getUint16(buf[off:]))
			off += 2
			tp.DownVol = uint32(getUint16(buf[off:]))
			off += 2
		case 3:
			if err = ckBufLen(off, 11, bLen, errS); err != nil {
				return
			}
			tp.Volume = getUint32(buf[off:])
			off += 4
			tp.UpVol = getUint32(buf[off:])
			off += 4
			tp.DownVol = getUint32(buf[off:])
			off += 4
		}
	}
	if idx != int(tickHd.Count) {
		err = errTickCount
		return
	}
	if off != bLen {
		err = errNoWay
	}
	return
}
