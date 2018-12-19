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
	if fd, err := os.Open(fileN); err != nil {
		return res, err
	} else if r, err := zlib.NewReader(fd); err != nil {
		return res, err
	} else {
		defer r.Close()
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return res, err
		}
		cnt := len(buf) / int(unsafe.Sizeof(tickDT{}))
		res = (*(*[1 << 31]tickDT)(unsafe.Pointer(&buf[0])))[1:cnt]
	}
	return
}

func LoadTickFX(pair string, startD, endD julian.JulianDay, cnt int) (res []TickFX, err error) {
	startD = startD.Weekbase()
	tCnt := 0
	var ticks []tickDT
	for startD < endD && tCnt < cnt {
		ticks, err = loadTickFX(pair, startD)
		if err != nil {
			return
		}
		tiCnt := len(ticks)
		rec := make([]TickFX, tiCnt)
		for i := 0; i < tiCnt; i++ {
			rec[i].Time, rec[i].Bid, rec[i].Ask = ticks[i].Time, ticks[i].Bid, ticks[i].Ask
		}
		res = append(res, rec...)
		tCnt += len(ticks)
		startD += 7
	}
	return
}

var errBufLen = errors.New("buf length dismatch sizeof minDT")

func loadMinFX(pair string, startD julian.JulianDay) (res []minDT, err error) {
	fileN := getMinPath(pair, startD)
	if fd, err := os.Open(fileN); err != nil {
		return res, err
	} else {
		defer fd.Close()
		buf, err := ioutil.ReadAll(fd)
		if err != nil {
			return res, err
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
				return res, err
			}
		}
	}
	return
}
