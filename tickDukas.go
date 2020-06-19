package ats

import (
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"os"
	"unsafe"

	"github.com/kjx98/golib/julian"
)

type tickDT struct {
	Time   DateTimeMs
	Ask    int32
	Bid    int32
	AskVol float32
	BidVol float32
}

type tNode struct {
	baseOff uint32
	ticks   []TickFX
	nextD   julian.JulianDay
}

type tickDB struct {
	pair    string
	startD  julian.JulianDay
	endD    julian.JulianDay
	cnt     int
	off     uint32
	curNode tNode
}

func getTickPath(pair string, startD julian.JulianDay) string {
	y, m, d := startD.Date()
	res := fmt.Sprintf("%s/forex/MinData/%s/%04d/%s-%04d%02d%02d.tic",
		homePath, pair, y, pair, y, m, d)
	return res
}

func loadTickFX(pair string, startD julian.JulianDay) (res []tickDT, err error) {
	fileN := getTickPath(pair, startD)
	if fd, errL := os.Open(fileN); errL != nil {
		err = errL
		return
	} else if r, errL := zlib.NewReader(fd); errL != nil {
		err = errL
		return
	} else {
		defer r.Close()
		if buf, errL := ioutil.ReadAll(r); errL == nil {
			cnt := len(buf) / int(unsafe.Sizeof(tickDT{}))
			res = (*(*[1 << 31]tickDT)(unsafe.Pointer(&buf[0])))[:cnt]
		} else {
			err = errL
			return
		}

	}
	return
}

// LoadTickFX ...	load DukasCopy forex tick data
//				startD, endD		0 unlimit, or weekbase of date
//				maxCnt				0 unlimit
func LoadTickFX(pair string, startD, endD julian.JulianDay, maxCnt int) (res []TickFX, err error) {
	if startD == 0 {
		if tiC, ok := initTicks[pair]; ok {
			startD = julian.FromUint32(tiC.TickStart)
			log.Info("LoadTickFX: startDate 0, replace to ", startD)
		}
	}
	startT := JulianToDateTimeMs(startD)
	startD = startD.Weekbase()
	tCnt := 0
	var ticks []tickDT
	for endD == 0 || startD < endD {
		ticks, err = loadTickFX(pair, startD)
		if err != nil {
			if len(res) > 0 && os.IsNotExist(err) {
				err = nil
			}
			return
		}
		tiCnt := len(ticks)
		if ticks[0].Time < startT {
			for i := 0; i < tiCnt; i++ {
				if ticks[i].Time < startT {
					continue
				}
				var rec TickFX
				rec.Time, rec.Bid, rec.Ask = ticks[i].Time, ticks[i].Bid, ticks[i].Ask
				res = append(res, rec)
				tCnt++
			}
		} else {
			rec := make([]TickFX, tiCnt)
			for i := 0; i < tiCnt; i++ {
				rec[i].Time, rec[i].Bid, rec[i].Ask = ticks[i].Time, ticks[i].Bid, ticks[i].Ask
			}
			res = append(res, rec...)
			tCnt += len(ticks)
		}
		if maxCnt != 0 && tCnt >= maxCnt {
			break
		}
		startD += 7
	}
	return
}
