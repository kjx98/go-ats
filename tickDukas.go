package ats

import (
	"compress/zlib"
	"fmt"
	"io"
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

func (sti *tickDB) TimeAt(i int) DateTimeMs {
	if i < 0 || i >= sti.cnt {
		panic("TimeAt out of bound")
	}
	// go to offset i
	if ccN := sti.nodeNum(i); ccN < 0 {
		return DateTimeMs(0)
	} else if ccN != sti.curNode {
		sti.curD = julian.JulianDay(7*ccN) + sti.startD
		sti.curNode = ccN
		sti.loadCurNode()
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
