package ats

import (
	"compress/zlib"
	"errors"
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

type minDT = MinFX

var homePath string
var noDukasData bool

func init() {
	homePath = os.Getenv("HOME")
	if ff, err := os.Stat(homePath + "/forex/MinData"); err != nil {
		noDukasData = true
	} else if !ff.IsDir() {
		noDukasData = true
	}
}

func getTickPath(pair string, startD julian.JulianDay) string {
	y, m, d := startD.Date()
	res := fmt.Sprintf("%s/forex/MinData/%s/%04d/%s-%04d%02d%02d.tic",
		homePath, pair, y, pair, y, m, d)
	return res
}

func getMinPath(pair string, startD julian.JulianDay) string {
	y, m, d := startD.Date()
	res := fmt.Sprintf("%s/forex/MinData/%s/%04d/%s-%04d%02d%02d.min",
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
		buf, errL := ioutil.ReadAll(r)
		if errL != nil {
			err = errL
			return
		}
		cnt := len(buf) / int(unsafe.Sizeof(tickDT{}))
		res = (*(*[1 << 31]tickDT)(unsafe.Pointer(&buf[0])))[1:cnt]
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

var errBufLen = errors.New("buf length dismatch sizeof minDT")

func loadMinFX(pair string, startD julian.JulianDay) (res []minDT, err error) {
	fileN := getMinPath(pair, startD)
	fd, errL := os.Open(fileN)
	if errL != nil {
		err = errL
		return
	}

	defer fd.Close()
	buf, errL := ioutil.ReadAll(fd)
	if errL != nil {
		err = errL
		return
	}
	cnt := len(buf) / 36
	if cnt > 0 {
		res = make([]minDT, cnt)
		j := 0
		for i := 0; i < cnt && j < len(buf); i++ {
			dst := (*(*[36]byte)(unsafe.Pointer(&res[i])))[:36]
			copy(dst, buf[j:j+36])
			j += 36
		}
		if j != len(buf) {
			err = errBufLen
			return
		}
	}
	return
}

var cacheDukasHits int
var cacheDukasMiss int

// DukasCacheStatus dump cache usage
func DukasCacheStatus() string {
	return fmt.Sprintf("DukasCache Status: Hits %d, Miss: %d", cacheDukasHits, cacheDukasMiss)
}

// LoadMinFX DukasCopy forex Min1 data
//		startD, endD		0 unlimit, or weekbase of date
//		maxCnt				0 unlimit
func LoadMinFX(pair string, startD, endD julian.JulianDay, maxCnt int) (res []MinFX, err error) {
	if startD == 0 {
		if tiC, ok := initTicks[pair]; ok {
			startD = julian.FromUint32(tiC.TickStart)
			log.Info("LoadMinFX: startDate 0, replace to ", startD)
		}
	}
	startD = startD.Weekbase()
	if cc, ok := cacheMinFX[pair]; ok {
		if startD >= cc.startD && endD == cc.endD {
			res = cc.res
			cacheDukasHits++
			return
		}
	}
	cacheDukasMiss++
	var cc = cacheMinFXType{startD: startD, endD: endD}

	tCnt := 0
	var mins []minDT
	for endD == 0 || startD < endD {
		mins, err = loadMinFX(pair, startD)
		if err != nil {
			if len(res) > 0 && os.IsNotExist(err) {
				err = nil
			}
			return
		}

		res = append(res, mins...)
		tCnt += len(mins)
		if maxCnt != 0 && tCnt >= maxCnt {
			break
		}
		startD += 7
	}
	cc.res = res
	if maxCnt == 0 {
		cacheMinFX[pair] = cc
	}
	return
}
