package ats

import (
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"unsafe"

	"github.com/kjx98/go-ats/lzma"
	"github.com/kjx98/golib/julian"
	"github.com/kjx98/golib/to"
)

// weekly combined forex DukasCopy Tick
type tickDT struct {
	Time   DateTimeMs
	Ask    int32
	Bid    int32
	AskVol float32
	BidVol float32
}

// raw lzma compressed Dukascopy Tick
type bi5TickDT struct {
	DeltaMs uint32
	Ask     int32
	Bid     int32
	AskVol  float32
	BidVol  float32
}

type tNode struct {
	baseOff uint32
	nTicks  uint32
}

type tickDB struct {
	pair    string
	startD  julian.JulianDay
	endD    julian.JulianDay
	curD    julian.JulianDay
	cnt     int
	curP    int
	curNode int
	tickBuf []TickFX
	nodes   []tNode
}

func getTickPath(pair string, startD julian.JulianDay) string {
	y, m, d := startD.Date()
	res := fmt.Sprintf("%s/forex/MinData/%s/%04d/%s-%04d%02d%02d.tic",
		homePath, pair, y, pair, y, m, d)
	return res
}

func getBi5Path(pair string, day julian.JulianDay, hr int) string {
	y, m, d := day.Date()
	res := fmt.Sprintf("%s/forex/DukasTick/%s/%04d/%02d/%02d/%02dh_ticks.bi5",
		homePath, pair, y, m-1, d, hr)
	return res
}

func getCurrencies() (res []string) {
	path := homePath + "/forex/DukasTick"
	if ff, err := ioutil.ReadDir(path); err == nil {
		for _, finfo := range ff {
			if !finfo.IsDir() {
				continue
			}
			if na := finfo.Name(); na[0] == '.' {
				log.Info("getCurrencies ignore ", na)
				continue
			} else {
				res = append(res, na)
			}
		}
	}
	return
}

func getDirMax(path string) (res string) {
	if ff, err := ioutil.ReadDir(path); err == nil {
		for _, finfo := range ff {
			if na := finfo.Name(); na[0] == '.' {
				log.Info("getCurrencies ignore ", na)
				continue
			} else if na > res {
				res = na
			}
		}
	}
	return
}

func checkLastTick(pair string) (day julian.JulianDay, hr int) {
	path := homePath + "/forex/DukasTick/" + pair
	ss := getDirMax(path)
	y := to.Int(ss)
	if y < 1900 {
		return
	}
	path += "/" + ss
	ss = getDirMax(path)
	m := to.Int(ss)
	if m < 0 {
		return
	}
	m++
	path += "/" + ss
	ss = getDirMax(path)
	d := to.Int(ss)
	if d < 0 {
		return
	}
	day = julian.NewJulianDay(y, m, d)
	path += "/" + ss
	ss = getDirMax(path)
	if len(ss) < 12 || ss[2:] != "h_ticks.bi5" {
		return
	}
	hr = to.Int(ss[:2])
	if hr < 0 {
		hr = 0
	} else {
		hr++
		if hr > 23 {
			hr = 0
			day = day.Add(1)
		}
	}
	return
}

func loadTickFX(pair string, startD julian.JulianDay) (res []tickDT, err error) {
	fileN := getTickPath(pair, startD)
	if fd, errL := os.Open(fileN); errL != nil {
		err = errL
		return
	} else if r, errL := zlib.NewReader(fd); errL != nil {
		fd.Close()
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

func rawI2F(v int32) float32 {
	rr := (*float32)(unsafe.Pointer(&v))
	return *rr
}

func loadBi5TickFX(pair string, startD julian.JulianDay, hr int) (res []tickDT, err error) {
	fileN := getBi5Path(pair, startD, hr)
	if fd, errL := os.Open(fileN); errL != nil {
		err = errL
		return
	} else {
		r := lzma.NewReader(fd)
		defer r.Close()
		var rv bi5TickDT
		startMS := JulianToDateTimeMs(startD).Add(hr * 3600 * 1000)
		for err == nil {
			err = binary.Read(r, binary.BigEndian, &rv)
			if err != nil {
				break
			}
			rsV := tickDT{Time: startMS.Add(int(rv.DeltaMs)),
				Ask: rv.Ask, Bid: rv.Bid,
				AskVol: rv.AskVol, BidVol: rv.BidVol}
			res = append(res, rsV)
		}
		if err == io.EOF {
			err = nil
			return
		}
	}
	return
}

func (sti *tickDB) Reset() {
	sti.curP = 0
	if sti.curNode != 0 {
		sti.curD = sti.startD
		sti.curNode = 0
		sti.loadCurNode()
	}
}

func (sti *tickDB) Len() int {
	return sti.cnt
}

func (sti *tickDB) Left() int {
	return sti.cnt - sti.curP
}

func (sti *tickDB) validateCur() int {
	curP := sti.curP
	if curP >= sti.cnt {
		panic("Out of tickDB bound")
	}
	if sti.curNode >= len(sti.nodes) {
		panic("Out of tickDB nodes bound")
	}
	ccN := sti.curNode
	ccOff := curP - int(sti.nodes[ccN].baseOff)
	if ccOff >= int(sti.nodes[ccN].nTicks) {
		panic("Out of tickDB curNode bound")
	}
	return ccOff
}

func (sti *tickDB) Time() DateTimeMs {
	curP := sti.validateCur()
	return sti.tickBuf[curP].Time
}

func (sti *tickDB) nodeNum(off int) int {
	if off < 0 || off >= sti.cnt {
		return -1
	}
	for idx, cNode := range sti.nodes {
		if off >= int(cNode.baseOff) && off < int(cNode.baseOff+cNode.nTicks) {
			return idx
		}
	}
	return -1
}

func (sti *tickDB) At(i int) error {
	if i < 0 || i >= sti.cnt {
		return errOutOfBound
	}
	// go to offset i
	if ccN := sti.nodeNum(i); ccN < 0 {
		return errOutOfBound
	} else if ccN != sti.curNode {
		sti.curD = julian.JulianDay(7*ccN) + sti.startD
		sti.curNode = ccN
		sti.loadCurNode()
	}
	return nil
}

func (sti *tickDB) TimeAt(i int) DateTimeMs {
	if sti.At(i) != nil {
		panic("TimeAt out of bound")
	}
	curP := sti.validateCur()
	return sti.tickBuf[curP].Time
}

func (sti *tickDB) TickValue() (bid, ask, last int32, vol uint32) {
	curP := sti.validateCur()
	return sti.tickBuf[curP].Bid, sti.tickBuf[curP].Ask, 0, 0
}

func (sti *tickDB) Next() error {
	sti.curP++
	ccN := sti.curNode
	if sti.curP < int(sti.nodes[ccN].baseOff+sti.nodes[ccN].nTicks) {
		return nil
	}
	ccN++
	if ccN >= len(sti.nodes) {
		return io.EOF
	}
	sti.curNode = ccN
	sti.curD += 7
	return sti.loadCurNode()
}

func (sti *tickDB) loadCurNode() error {
	ticks, err := loadTickFX(sti.pair, sti.curD)
	if err != nil {
		return err
	}
	tiCnt := len(ticks)
	rec := make([]TickFX, tiCnt)
	for i := 0; i < tiCnt; i++ {
		rec[i].Time, rec[i].Bid, rec[i].Ask = ticks[i].Time, ticks[i].Bid, ticks[i].Ask
	}
	sti.tickBuf = rec
	return nil
}

var tickDbMap = map[string]tickDB{}

// OpenTickFX		Open DukasCopy forex tick data
//				startD, endD		0 unlimit, or weekbase of date
//				maxCnt				0 unlimit
func OpenTickFX(pair string, startD, endD julian.JulianDay, maxCnt int) (res *tickDB, err error) {
	if startD == 0 {
		if tiC, ok := initTicks[pair]; ok {
			startD = julian.FromUint32(tiC.TickStart)
			log.Info("LoadTickFX: startDate 0, replace to ", startD)
		}
	}
	startD = startD.Weekbase()
	tDB := tickDbMap[pair]
	if tDB.pair == pair && tDB.startD <= startD && tDB.endD >= endD {
		res = &tDB
		log.Info("OpenTickFX using cache, startD ", startD)
		return
	}
	tCnt := 0
	tOff := 0
	var ticks []tickDT
	res = &tickDB{
		pair:   pair,
		startD: startD,
		endD:   endD,
		curD:   startD,
	}
	for endD == 0 || startD < endD {
		ticks, err = loadTickFX(pair, startD)
		if err != nil {
			if res.cnt > 0 && os.IsNotExist(err) {
				err = nil
			}
			if err == nil {
				tickDbMap[pair] = *res
			}
			return
		}
		tiCnt := len(ticks)
		cNode := tNode{
			baseOff: uint32(tOff),
			nTicks:  uint32(tiCnt),
		}
		tOff += tiCnt
		res.cnt += tiCnt
		res.nodes = append(res.nodes, cNode)
		tCnt += tiCnt
		if startD == res.curD {
			rec := make([]TickFX, tiCnt)
			for i := 0; i < tiCnt; i++ {
				rec[i].Time, rec[i].Bid, rec[i].Ask = ticks[i].Time, ticks[i].Bid, ticks[i].Ask
			}
			res.tickBuf = rec
		}
		if maxCnt != 0 && tCnt >= maxCnt {
			break
		}
		startD += 7
	}
	tickDbMap[pair] = *res
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
