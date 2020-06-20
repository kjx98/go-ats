package ats

import (
	"encoding/csv"
	"errors"
	"io"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kjx98/avl"
	"github.com/kjx98/golib/julian"
	"github.com/kjx98/golib/to"
)

type account struct {
	fundStart  float64
	equity     float64
	balance    float64
	fund       float64
	margin     float64 // freeze fund for margin
	trades     int
	winTrades  int
	lossTrades int
	profit     float64
	loss       float64

	evChan chan<- QuoteEvent
	orders []int
	pos    map[SymbolKey]*PositionType
}

type simOrderType struct {
	simBroker
	oid   int
	price int32
	OrderType
}

type orderBook struct {
	bids, asks *avl.Tree
}

type simTick struct {
	curP  int
	ticks []Tick
}

type simTickFX struct {
	curP  int
	ticks []TickFX
}

func (sti *simTick) Reset() {
	sti.curP = 0
}

func (sti *simTick) Len() int {
	return len(sti.ticks)
}

func (sti *simTick) Left() int {
	return len(sti.ticks) - sti.curP
}

func (sti *simTick) Time() DateTimeMs {
	curP := sti.curP
	if curP >= len(sti.ticks) {
		panic("Out of simTick bound")
	}
	return sti.ticks[curP].Time.DateTimeMs()
}

func (sti *simTick) TimeAt(i int) DateTimeMs {
	if i >= len(sti.ticks) {
		panic("Out of simTick bound")
	}
	return sti.ticks[i].Time.DateTimeMs()
}

func (sti *simTick) TickValue() (bid, ask, last int32, vol uint32) {
	curP := sti.curP
	if curP >= len(sti.ticks) {
		panic("Out of simTick bound")
	}
	return 0, 0, sti.ticks[curP].Last, sti.ticks[curP].Volume
}

func (sti *simTick) Next() error {
	sti.curP++
	if sti.curP >= len(sti.ticks) {
		return io.EOF
	}
	return nil
}

func (sti *simTickFX) Reset() {
	sti.curP = 0
}

func (sti *simTickFX) Len() int {
	return len(sti.ticks)
}

func (sti *simTickFX) Left() int {
	return len(sti.ticks) - sti.curP
}

func (sti *simTickFX) Time() DateTimeMs {
	curP := sti.curP
	if curP >= len(sti.ticks) {
		panic("Out of simTick bound")
	}
	return sti.ticks[curP].Time
}

func (sti *simTickFX) TimeAt(i int) DateTimeMs {
	if i >= len(sti.ticks) {
		panic("Out of simTick bound")
	}
	return sti.ticks[i].Time
}

func (sti *simTickFX) TickValue() (bid, ask, last int32, vol uint32) {
	curP := sti.curP
	if curP >= len(sti.ticks) {
		panic("Out of simTick bound")
	}
	return sti.ticks[curP].Bid, sti.ticks[curP].Ask, 0, 0
}

func (sti *simTickFX) Next() error {
	sti.curP++
	if sti.curP >= len(sti.ticks) {
		return io.EOF
	}
	return nil
}

type simTicker interface {
	Reset()
	Len() int
	Left() int
	Time() DateTimeMs
	TimeAt(i int) DateTimeMs
	Next() error
	TickValue() (bid, ask, last int32, vol uint32)
}

func bidCompare(a, b interface{}) int {
	ora, ok := a.(*simOrderType)
	// maybe panic, if not simOrderType
	if !ok {
		return 0
	}
	orb, ok := b.(*simOrderType)
	if !ok {
		return 0
	}
	if ora.price == orb.price {
		return ora.oid - orb.oid
	}
	if ora.price == 0 {
		return -1
	}
	if orb.price == 0 {
		return 1
	}
	// low price, low priority
	return int(orb.price) - int(ora.price)
}

func askCompare(a, b interface{}) int {
	ora, ok := a.(*simOrderType)
	// maybe panic, if not simOrderType
	if !ok {
		return 0
	}
	orb, ok := b.(*simOrderType)
	if !ok {
		return 0
	}
	if ora.price == orb.price {
		return ora.oid - orb.oid
	}
	// high price, low priority
	return int(ora.price) - int(orb.price)
}

var acctLock sync.RWMutex
var nAccounts int
var simAccounts = map[simBroker]*account{}
var orderLock sync.RWMutex
var orderNo int
var simOrders = map[int]*simOrderType{}
var bSimValidate = true

var defaultFund = float64(1e6)

// simulate params
// sim Run start/stop time
var startTime, endTime timeT64
var simPeriod Period

// current time DateTimeMs of sim Run VM
var simCurrent DateTimeMs
var simVmLock sync.RWMutex
var simTickMap = map[SymbolKey]simTicker{}
var simTickRun map[SymbolKey]simTicker

// simStatus should be atomic
var simStatus int32

//var maxAllocHeap uint64
//var maxSysHeap uint64
//var timeAtMaxAlloc DateTimeMs
var onceLoad sync.Once

// simSymbolQ symbol fKey map
var simSymbolsQ = map[SymbolKey]*Quotes{}

// orderBook map with symbol key
var simOrderBook = map[string]orderBook{}

// VmIdle ... vm is idle
const (
	VmIdle int32 = iota
	VmStart
	VmRunning
	VmStoping
)

var (
	errVMStatus     = errors.New("simBroker VM status error")
	errTickNonExist = errors.New("Tick Data not exist")
	errTickOrder    = errors.New("Tick Data order error")
	errNoOrder      = errors.New("No such order")
	errCancelOrder  = errors.New("can't cancel,canceled or filled")
)

// InitSimBroker ... set default fund, startTime, endTime  etc
func InitSimBroker(startT, endT time.Time, defFund float64) {
	startTime = timeT64FromTime(startT)
	endTime = timeT64FromTime(endT)
	if defFund > 10000 {
		defaultFund = defFund
	}
}

func simInsertOrder(or *simOrderType) {
	orBook, ok := simOrderBook[or.Symbol]
	if !ok {
		orBook.bids = avl.New(bidCompare)
		orBook.asks = avl.New(askCompare)
		simOrderBook[or.Symbol] = orBook
	}
	if or.OrderType.Dir.Sign() > 0 {
		// bid
		orBook.bids.Insert(or)
	} else {
		orBook.asks.Insert(or)
	}
}

func simRemoveOrder(or *simOrderType) {
	if orBook, ok := simOrderBook[or.Symbol]; ok {
		if or.OrderType.Dir.Sign() > 0 {
			if v := orBook.bids.Find(or); v != nil {
				orBook.bids.Remove(v)
			}
		} else {
			if v := orBook.asks.Find(or); v != nil {
				orBook.asks.Remove(v)
			}
		}
	}
}

func dumpSimOrderBook(sym string) {
	orB, ok := simOrderBook[sym]
	if !ok {
		log.Info("no OrderBook for ", sym)
		return
	}
	log.Infof("Dump %s bids:", sym)
	iter := orB.bids.Iterator(avl.Forward)
	for node := iter.First(); node != nil; node = iter.Next() {
		v := node.Value.(*simOrderType)
		log.Infof(" No:%d %s %d %s %g %d", v.oid, v.Symbol, v.price,
			v.OrderType.Dir, v.Price, v.Qty)
	}
	log.Infof("Dump %s asks:", sym)
	iter = orB.asks.Iterator(avl.Forward)
	for node := iter.First(); node != nil; node = iter.Next() {
		v := node.Value.(*simOrderType)
		log.Infof(" No:%d %s %d %s %g %d", v.oid, v.Symbol, v.price,
			v.OrderType.Dir, v.Price, v.Qty)
	}

}

func dumpSimOrderStats() {
	totalOrders := 0
	for sym, orB := range simOrderBook {
		log.Infof("%s Bid orders: %d, Ask orders: %d", sym, orB.bids.Len(), orB.asks.Len())
		totalOrders += orB.bids.Len() + orB.asks.Len()
	}
	log.Infof("Total unfilled orders: %d", totalOrders)
}

func dumpSimBroker() {
	for k, acct := range simAccounts {
		if len(acct.orders) == 0 {
			continue
		}
		log.Infof("SimBroker(%d) fundStart(%g) end Fund(%g) trades(%d of %d) "+
			"win/loss(%d/%d) Profit/Loss(%.3f/%.3f)", int(k), acct.fundStart, acct.fund,
			acct.trades, len(acct.orders), acct.winTrades, acct.lossTrades,
			acct.profit, acct.loss)
		for fk, pp := range acct.pos {
			si, _ := fk.SymbolInfo()
			log.Infof("simBroker(%d) position(%s) %d avrPrice(%.3f)", int(k), si.Ticker,
				pp.Positions, pp.AvgPrice)
		}
	}
}

type simBroker int

func (b simBroker) Open(ch chan<- QuoteEvent) (Broker, error) {
	acctLock.Lock()
	defer acctLock.Unlock()

	var acct = account{fundStart: defaultFund, fund: defaultFund, evChan: ch,
		equity: defaultFund, balance: defaultFund}
	acct.orders = []int{}
	acct.pos = map[SymbolKey]*PositionType{}
	nAccounts++
	//log.Info("dump acct:", ch, nAccounts, acct)
	//simAccounts is map
	bb := simBroker(nAccounts)

	simAccounts[bb] = &acct
	return bb, nil
}

func simLoadSymbols() {
	onceLoad.Do(func() {
		if fd, err := os.Open("universe.csv"); err != nil {
			panic("open universe.csv error")
		} else {
			defer fd.Close()
			csvR := csv.NewReader(fd)
			simPeriod = Daily
			line, err := csvR.Read()
			for err == nil {
				// process a line
				if len(line) == 0 || line[0] == "" {
					continue
				}
				newSymbolInfo(line[0])
				if si, err := GetSymbolInfo(line[0]); err == nil {
					// try load tick, min, day data
					var st, dt julian.JulianDay
					var bNeedForge = true
					if len(line) > 2 {
						st = julian.FromUint32(uint32(to.Int(line[2])))
						if len(line) > 3 {
							dt = julian.FromUint32(uint32(to.Int(line[3])))
						}
					}
					if si.IsForex {
						if strings.Contains(line[1], "t") {
							if res, err := OpenTickFX(line[0], st, dt, 0); err == nil {
								// load to sim
								simTickMap[si.FastKey()] = res
								bNeedForge = false
								// no OnTick right now
								/*
									if simPeriod > 0 {
										simPeriod = 0
									}
								*/
							}
							/*
								if res, err := LoadTickFX(line[0], st, dt, 0); err == nil {
									// load to sim
									var tickD = simTickFX{}
									tickD.ticks = res
									simTickMap[si.FastKey()] = &tickD
									bNeedForge = false
									// no OnTick right now
								}
							*/
						}
					} else {
						if strings.Contains(line[1], "t") {
							// try load ticks for Non FX
						}
					}
					if strings.Contains(line[1], "m") {
						// loadMindata
						if si.IsForex {
							if err := LoadBarFX(line[0], Min1, st, dt); err == nil {
								if simPeriod > Min1 {
									simPeriod = Min1
								}
							}
						} else {
							// try load Min5
						}
					}
					if strings.Contains(line[1], "d") {
						// load daily Bar
						if si.IsForex {
							// load FX daily
						} else {
							LoadDayBar(line[0], Daily, st, dt)
						}
					}
					if bNeedForge {
						// no tick, forge tick from Min1/Min5 or Daily
						forgeTicks(&si)
					}
				}
				line, err = csvR.Read()
			}
		}
	})
	// rebuild simTickRun
	if len(simTickRun) > 0 {
		for k := range simTickRun {
			delete(simTickRun, k)
		}
	}
	simTickRun = map[SymbolKey]simTicker{}
	for k, v := range simTickMap {
		simTickRun[k] = v
	}
}

func forgeTicks(si *SymbolInfo) {
	forgeTicksFromBar := func(cc cacheTAer, period Period) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		if si.IsForex {
			var tickD = simTickFX{}
			tickD.ticks = make([]TickFX, cc.Len()*4)
			j := 0
			for i := 0; i < cc.Len(); i++ {
				var atick *TickFX
				ti, o, h, l, c, _ := cc.BarValue(i)
				atick = &tickD.ticks[j]
				j++
				atick.Time = ti.DateTimeMs()
				atick.Bid = o
				atick.Ask = o + si.DefSpread
				atick = &tickD.ticks[j]
				j++
				hto := r.Int63() % int64(period)
				lto := r.Int63() % int64(period)
				if hto > lto {
					atick.Time = (ti + timeT64(lto)).DateTimeMs()
					atick.Bid = l
					atick.Ask = l + si.DefSpread
					atick = &tickD.ticks[j]
					j++
					atick.Time = (ti + timeT64(hto)).DateTimeMs()
					atick.Bid = h
					atick.Ask = h + si.DefSpread
				} else {
					atick.Time = (ti + timeT64(hto)).DateTimeMs()
					atick.Bid = h
					atick.Ask = h + si.DefSpread
					atick = &tickD.ticks[j]
					j++
					atick.Time = (ti + timeT64(lto)).DateTimeMs()
					atick.Bid = l
					atick.Ask = l + si.DefSpread
				}
				// last for close
				atick = &tickD.ticks[j]
				j++
				atick.Time = (ti + timeT64(period)).DateTimeMs() - 1
				atick.Bid = c
				atick.Ask = c + si.DefSpread
			}
			simTickMap[si.FastKey()] = &tickD
		} else {
			var tickD = simTick{}
			tickD.ticks = make([]Tick, cc.Len()*4)
			j := 0
			for i := 0; i < cc.Len(); i++ {
				var atick *Tick
				ti, o, h, l, c, vol := cc.BarValue(i)
				atick = &tickD.ticks[j]
				j++
				atick.Time = timeT32(ti)
				atick.Last = o
				atick.Volume = uint32(vol * 3 / 8)
				atick = &tickD.ticks[j]
				j++
				hto := r.Int63() % int64(period)
				lto := r.Int63() % int64(period)
				if hto > lto {
					atick.Time = timeT32(ti + timeT64(lto))
					atick.Last = l
					atick.Volume = uint32(vol / 8)
					atick = &tickD.ticks[j]
					j++
					atick.Time = timeT32(ti + timeT64(hto))
					atick.Last = h
					atick.Volume = uint32(vol / 8)
				} else {
					atick.Time = timeT32(ti + timeT64(hto))
					atick.Last = h
					atick.Volume = uint32(vol / 8)
					atick = &tickD.ticks[j]
					j++
					atick.Time = timeT32(ti + timeT64(lto))
					atick.Last = l
					atick.Volume = uint32(vol / 8)
				}
				// last for close
				atick = &tickD.ticks[j]
				j++
				atick.Time = timeT32(ti+timeT64(period)) - 1
				atick.Last = c
				atick.Volume = uint32(vol * 3 / 8)
			}
			simTickMap[si.FastKey()] = &tickD
		}
	}
	if si.IsForex {
		if cc, ok := cacheMinFX[si.Ticker]; ok {
			// forge via FX Min1
			forgeTicksFromBar(&cc, Min1)
			return
		}
	} else {
		if cc, ok := cacheMinTA[si.Ticker]; ok {
			// forge via Min5
			forgeTicksFromBar(&cc, Min5)
			return
		}
	}
	if cc, ok := cacheDayTA[si.Ticker]; ok {
		// forge via Daily
		forgeTicksFromBar(&cc, Daily)
		return
	}
}

// loadRunTick ... load Tick data from simTickRun
func loadRunTick(sym string) (simTicker, error) {
	if si, err := GetSymbolInfo(sym); err != nil {
		return nil, err
	} else if v, ok := simTickRun[si.FastKey()]; ok {
		return v, nil
	}
	return nil, errTickNonExist
}

// ValidateTick ... validate tick data timestamp in order
func ValidateTick(sym string) error {
	if si, err := GetSymbolInfo(sym); err != nil {
		return err
	} else if v, ok := simTickMap[si.FastKey()]; ok {
		var oldTi DateTimeMs
		var min, max int32
		defer v.Reset()
		for i := 0; i < v.Len(); i++ {
			if ti := v.Time(); ti >= oldTi {
				oldTi = ti
			} else {
				return errTickOrder
			}
			if si.IsForex {
				bid, ask, _, _ := v.TickValue()
				if min == 0 || bid < min {
					min = bid
				}
				if ask > max {
					max = ask
				}
			}
			v.Next()
		}
		log.Infof("%s ticks ok, min: %d, max: %d", si.Ticker, min, max)
	} else {
		return errTickNonExist
	}
	return nil
}

// every instance of VM should be with same configure
func (b simBroker) Start(c Config) error {
	// read Config, ...
	// start goroutine for simulate/backtesting
	switch atomic.LoadInt32(&simStatus) {
	case VmIdle:
	case VmStart, VmRunning:
		return nil
	default:
		return errVMStatus
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	//maxAllocHeap = 0
	//maxSysHeap = 0
	//timeAtMaxAlloc = 0
	atomic.StoreInt32(&simStatus, VmStart)
	// load Bars
	// build ticks
	simLoadSymbols()
	startMs := DateTimeMs(0)
	msStart := startTime.DateTimeMs()
	//msEnd := endTime.DateTimeMs()
	// try get the first tick datetime
	for k, v := range simTickRun {
		// skip to msStart
		for i := 0; i < v.Len(); i++ {
			if v.Time() >= msStart {
				if startMs == 0 || v.Time() < startMs {
					startMs = v.Time()
				}
				break
			}
			if v.Next() != nil {
				break
			}
		}
		si, _ := k.SymbolInfo()
		if v.Time() < msStart {
			log.Infof("delete simTickRun for symbol(%s)", si.Ticker)
			delete(simTickRun, k)
		} else {
			log.Infof("symbol(%s) left %d ticks", si.Ticker, v.Left())
		}
	}

	// start Tick feed goroutine
	simCurrent = startMs
	// always run tick
	/*
		if c.GetInt("RunTick", 0) != 0 {
			// go routint tick matcher
		} else {
			// go bar matcher
		}
	*/
	atomic.StoreInt32(&simStatus, VmRunning)
	// start go routine process ticks
	go simDoTickLoop()
	return nil
}

func simDoTickLoop() {
	startT := time.Now()
	totalTicks := 0
	totalDays := 0
	var msEnd DateTimeMs
	if endTime != 0 {
		msEnd = endTime.DateTimeMs()
	}
	nextPeriod, _ := periodBaseTime(simCurrent.Unix(), simPeriod)
	nextPeriod += int64(simPeriod)
	nextDay, _ := periodBaseTime(simCurrent.Unix(), Daily)
	nextDay += int64(Daily)
	if len(simTickRun) == 0 {
		log.Info("Empty simTickRun, status to Idle")
	} else {
		log.Info("simStart:", simCurrent, " --> simEnd:", msEnd)
		log.Info("number of Subscribed quote:", len(simSymbolsQ))
	}
	for len(simTickRun) > 0 && atomic.LoadInt32(&simStatus) == VmRunning {
		msNext := DateTimeMs(0)
		for k, v := range simTickRun {
			var ticker string
			var si *SymbolInfo
			if s, err := k.SymbolInfo(); err != nil {
				delete(simTickRun, k)
				continue
			} else {
				si = s
				ticker = si.Ticker
			}
			if v.Time() == simCurrent {
				totalTicks++
				// should update quote & Bars
				simUpdateQuote(si, v)
				// shall emit Min1/Min5 event?
				// process OrderBook
				simMatchOrder(si, v)
				// emit a tick
				//simEmitEvent(QuoteEvent{Symbol: ticker, EventID: 0})
				// move to next
				if err := v.Next(); err != nil {
					log.Infof("delete simTickRun for symbol(%s) EOF", ticker)
					delete(simTickRun, k)
					continue
				}
			}
			if msNext == 0 || v.Time() < msNext {
				msNext = v.Time()
			}

		}
		simCurrent = msNext
		if simCur := simCurrent.Unix(); simCur >= nextPeriod {
			nextPeriod, _ = periodBaseTime(simCur, simPeriod)
			nextPeriod += int64(simPeriod)
			simEmitEvents(QuoteEvent{EventID: int(simPeriod)})
			if simCur > nextDay {
				totalDays++
				simDayRotate()
				nextDay, _ = periodBaseTime(simCur, Daily)
				nextDay += int64(Daily)
				simEmitEvents(QuoteEvent{EventID: int(Daily)})
			}
		}
		if msEnd != 0 && msNext > msEnd {
			break
		}
	}
	// emit run out of tick
	simEmitEvents(QuoteEvent{EventID: -1})
	// clean simTickRun for manual stop
	if len(simTickRun) > 0 {
		log.Info("MANUAL stop simDoTickLoop")
		for k := range simTickRun {
			delete(simTickRun, k)
		}
	}
	atomic.StoreInt32(&simStatus, VmIdle)
	endT := time.Now()
	durT := endT.Sub(startT).Seconds()
	log.Infof("simDoTickLoop run %d ticks %d Days cost %.3f seconds, %.3g TPS",
		totalTicks, totalDays, durT, float64(totalTicks)/durT)
}

func simUpdateQuote(si *SymbolInfo, tick simTicker) {
	if qq, ok := simSymbolsQ[si.FastKey()]; ok {
		qq.UpdateTime = simCurrent
		bid, ask, last, vol := tick.TickValue()
		if si.IsForex {
			last = bid
		}
		fBid := float64(bid) * si.Divi()
		fAsk := float64(ask) * si.Divi()
		fLast := float64(last) * si.Divi()
		qq.Bid, qq.Ask, qq.Last = fBid, fAsk, fLast
		qq.Volume += int64(vol)
		if qq.TodayOpen == 0 {
			qq.TodayOpen = fLast
		}
		if qq.TodayHigh < fLast {
			qq.TodayHigh = fLast
		}
		if qq.TodayLow == 0 || qq.TodayLow > fLast {
			qq.TodayLow = fLast
		}
	}
}

func simUpdateAcctPos(si *SymbolInfo, or *simOrderType, last int32, vol int) (profit float64) {
	if vol <= 0 {
		return
	}
	acct := simAccounts[or.simBroker]
	acct.trades++
	var pos *PositionType
	if po, ok := acct.pos[si.FastKey()]; ok {
		pos = po
	} else {
		pos = &PositionType{fKey: si.FastKey()}
		acct.pos[si.FastKey()] = pos
	}
	fLast := float64(last) * si.Divi()
	switch or.Dir.Sign() {
	case 1: // for buy
		if pos.Positions >= 0 {
			// increase position
			avg := pos.AvgPrice*float64(pos.Positions) + fLast*float64(vol)
			pos.Positions += vol
			pos.AvgPrice = avg / float64(pos.Positions)
		} else {
			// close offset
			profit = si.CalcProfit(pos.AvgPrice, fLast, -vol)
			pos.Positions += vol
			acct.fund += profit
			acct.balance += profit
			if profit >= 0 {
				acct.profit += profit
				acct.winTrades++
			} else {
				acct.loss += profit
				acct.lossTrades++
			}
		}
	case -1: // for sell
		if pos.Positions <= 0 {
			// increase position
			avg := pos.AvgPrice*float64(-pos.Positions) + fLast*float64(vol)
			pos.Positions -= vol
			pos.AvgPrice = avg / float64(-pos.Positions)
		} else {
			// close offset
			profit = si.CalcProfit(pos.AvgPrice, fLast, vol)
			pos.Positions -= vol
			acct.fund += profit
			acct.balance += profit
			if profit >= 0 {
				acct.profit += profit
				acct.winTrades++
			} else {
				acct.loss += profit
				acct.lossTrades++
			}
		}
	default:
		// should be error
	}
	return
}

var simLogMatchs int

func simMatchOrder(si *SymbolInfo, tick simTicker) {
	setFill := func(or *simOrderType, last int32, vol int) {
		or.OrderType.QtyFilled = or.OrderType.Qty
		vol = or.OrderType.Qty
		or.OrderType.Status = OrderFilled
		or.DoneTime = simCurrent
		pl := simUpdateAcctPos(si, or, last, vol)
		simLogMatchs++
		if simLogMatchs <= 10 {
			log.Infof("Filled No:%d %s %d %s %g %d P&L(%.3f) via broker(%d)", or.oid, or.Symbol,
				or.price, or.Dir, or.Price, or.Qty, pl, int(or.simBroker))
		}
	}
	if orB, ok := simOrderBook[si.Ticker]; ok {
		bid, ask, last, vol := tick.TickValue()
		if si.IsForex {
			last = ask
		}
		iter := orB.bids.Iterator(avl.Forward)
		for node := iter.First(); node != nil; node = iter.Next() {
			v := node.Value.(*simOrderType)
			if v.price >= last {
				// match
				setFill(v, last, int(vol))
				orB.bids.Remove(node)
			} else {
				break
			}
		}
		if si.IsForex {
			last = bid
		}
		iter = orB.asks.Iterator(avl.Forward)
		for node := iter.First(); node != nil; node = iter.Next() {
			v := node.Value.(*simOrderType)
			if v.price <= last {
				// match
				setFill(v, last, int(vol))
				orB.asks.Remove(node)
			} else {
				break
			}
		}
	}
}

func simEmitOneEvent(ev QuoteEvent) {
	sendEvent := func(ch chan<- QuoteEvent) {
		if ch == nil {
			return
		}
		for {
			select {
			case ch <- ev:
				return
			default:
				runtime.Gosched()
				// inscrease block count
			}
		}
	}
	for _, bb := range simAccounts {
		if bb.evChan == nil {
			continue
		}
		sendEvent(bb.evChan)
	}
}

func simDayRotate() {
	for fk, qq := range simSymbolsQ {
		la := qq.Last
		*qq = Quotes{}
		qq.Pclose = la
		qq.UpdateTime = simCurrent
		if si, err := fk.SymbolInfo(); err == nil {
			var ev = QuoteEvent{Symbol: si.Ticker, EventID: int(Daily)}
			// emit event
			simEmitOneEvent(ev)
		}
	}
}

func simEmitEvents(ev QuoteEvent) {
	if ev.Symbol != "" {
		// emit one event
		simEmitOneEvent(ev)
		return
	}
	for fk := range simSymbolsQ {
		if si, err := fk.SymbolInfo(); err == nil {
			ev.Symbol = si.Ticker
			// emit event
			simEmitOneEvent(ev)
		}
	}
}

func (b simBroker) Stop() error {
	switch atomic.LoadInt32(&simStatus) {
	case VmIdle, VmStoping:
		return nil
	case VmRunning:
	default:
		return errVMStatus
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	atomic.StoreInt32(&simStatus, VmStoping)
	// stop Bar feed
	atomic.StoreInt32(&simStatus, VmIdle)
	return nil
}

func (b simBroker) SubscribeQuotes(qq []QuoteSubT) error {
	if atomic.LoadInt32(&simStatus) != VmIdle {
		return errVMStatus
	}
	// prepare Bars
	// maybe Once load?
	simVmLock.Lock()
	defer simVmLock.Unlock()
	// update QuotesPtr only not subscribed
	for _, qs := range qq {
		if si, err := GetSymbolInfo(qs.Symbol); err != nil {
			continue
		} else {
			if _, ok := simSymbolsQ[si.fKey]; !ok {
				simSymbolsQ[si.fKey] = qs.QuotesPtr
			}
		}
	}
	return nil
}

func (b simBroker) Equity() float64 {
	acct := simAccounts[b]
	return acct.equity
}

func (b simBroker) Balance() float64 {
	acct := simAccounts[b]
	return acct.balance
}

func (b simBroker) Cash() float64 {
	acct := simAccounts[b]
	return acct.fund
}

func (b simBroker) FreeMargin() float64 {
	acct := simAccounts[b]
	return acct.equity - acct.margin
}

func (b simBroker) SendOrder(sym string, dir OrderDirT, qty int, prc float64,
	stopL float64) int {
	si, err := GetSymbolInfo(sym)
	if err != nil {
		return -1
	}
	var prcI = int32(prc * si.Multi())
	// tobe fix
	// verify, put to orderbook
	if bSimValidate {
		// validate order margin
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	orderNo++
	var or = simOrderType{simBroker: b, oid: orderNo, price: prcI,
		OrderType: OrderType{Symbol: sym, Price: prc, StopPrice: stopL,
			Dir: dir, Qty: qty}}
	or.AckTime = simCurrent
	simOrders[orderNo] = &or
	// put to orderBook
	simInsertOrder(&or)
	acct := simAccounts[b]
	acct.orders = append(acct.orders, orderNo)
	return orderNo
}

func simOrderInAcct(acct *account, oid int) bool {
	idx := sort.SearchInts(acct.orders, oid)
	if idx >= len(acct.orders) || acct.orders[idx] != oid {
		return false
	}
	return true
}

func (b simBroker) CancelOrder(oid int) error {
	acct := simAccounts[b]
	if oid > orderNo {
		return errNoOrder
	}
	if !simOrderInAcct(acct, oid) {
		// no such order
		return errNoOrder
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	// remove order from orderbook
	or, ok := simOrders[oid]
	if !ok {
		return errNoOrder
	}
	switch or.OrderType.Status {
	case OrderFilled, OrderCanceled:
		return errCancelOrder
	default:
		simRemoveOrder(or)
		or.OrderType.Status = OrderCanceled
		or.DoneTime = simCurrent
	}
	return nil
}

func (b simBroker) CloseOrder(oId int) {
	acct := simAccounts[b]
	// if open, close with market
	// if stoploss, remove stoploss, change to market
	if oId > orderNo {
		return
	}
	if !simOrderInAcct(acct, oId) {
		// no such order
		return
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	// if order open or partfill, changed to market order
	// remove order from orderbook
	or, ok := simOrders[oId]
	if !ok {
		return
	}
	simRemoveOrder(or)
	switch or.OrderType.Status {
	case OrderAccept, OrderPartFilled:
		// change to market order
		if or.OrderType.StopPrice != 0 {
			or.OrderType.StopPrice = 0
		}
		or.OrderType.Price = 0
		or.price = 0
		simInsertOrder(or)
	case OrderFilled:
		// do nothing
	default:
		or.OrderType.Status = OrderCanceled
		or.DoneTime = simCurrent
	}
}

func (b simBroker) GetOrder(oId int) *OrderType {
	orderLock.RLock()
	defer orderLock.RUnlock()
	if o, ok := simOrders[oId]; ok {
		return &o.OrderType
	}
	return nil
}

func (b simBroker) GetOrders() []int {
	acct := simAccounts[b]
	return acct.orders
}

func (b simBroker) GetPosition(sym string) (vPos PositionType) {
	si, err := GetSymbolInfo(sym)
	if err != nil {
		return
	}
	acct := simAccounts[b]
	if v, ok := acct.pos[si.fKey]; ok {
		vPos = *v
	}
	return
}

func (b simBroker) GetPositions() (res []PositionType) {
	acct := simAccounts[b]
	if len(acct.pos) == 0 {
		return
	}
	res = make([]PositionType, len(acct.pos))
	i := 0
	for _, v := range acct.pos {
		res[i] = *v
		i++
	}
	return
}

//go:noinline
func (b simBroker) TimeCurrent() DateTimeMs {
	return simCurrent
}

var simTrader simBroker

func init() {
	RegisterBroker("simBroker", simTrader)
}
