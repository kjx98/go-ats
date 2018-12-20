package ats

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

func LoadTickFX(pair string, startD, endD julian.JulianDay, cnt int) (res []TickFX, err error) {
	if startD == 0 {
		if tiC, ok := initTicks[pair]; ok {
			startD = julian.FromUint32(tiC.TickStart)
			log.Println("LoadTickFX: startDate 0, replace to ", startD)
		}
	}
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
		rec := make([]TickFX, tiCnt)
		for i := 0; i < tiCnt; i++ {
			rec[i].Time, rec[i].Bid, rec[i].Ask = ticks[i].Time, ticks[i].Bid, ticks[i].Ask
		}
		res = append(res, rec...)
		tCnt += len(ticks)
		if tCnt >= cnt {
			break
		}
		startD += 7
	}
	return
}

var errBufLen = errors.New("buf length dismatch sizeof minDT")

func loadMinFX(pair string, startD julian.JulianDay) (res []minDT, err error) {
	fileN := getMinPath(pair, startD)
	if fd, errL := os.Open(fileN); errL != nil {
		err = errL
		return
	} else {
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
	}
	return
}

func LoadMinFX(pair string, startD, endD julian.JulianDay, cnt int) (res []MinFX, err error) {
	if startD == 0 {
		if tiC, ok := initTicks[pair]; ok {
			startD = julian.FromUint32(tiC.TickStart)
			log.Println("LoadMinFX: startDate 0, replace to ", startD)
		}
	}
	startD = startD.Weekbase()
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
		if tCnt >= cnt {
			break
		}
		startD += 7
	}
	return
}
