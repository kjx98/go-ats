package ats

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/kjx98/avl"
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
	pos    []PositionType
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
	// low price, low priority
	return int(ora.price) - int(orb.price)
}

var acctLock sync.RWMutex
var nAccounts int
var simAccounts = []*account{}
var orderLock sync.RWMutex
var nOrders int
var simOrders = []simOrderType{}
var startTime, endTime timeT64
var simCurrent DateTimeMs
var simVmLock sync.RWMutex

// simStatus should be atomic
var simStatus int32
var maxAllocHeap uint64
var maxSysHeap uint64
var timeAtMaxAlloc DateTimeMs

// simSymbolQ symbol fKey map
var simSymbolsQ = map[int]*Quotes{}

// orderBook map with symbol key
var simOrderBook map[string]orderBook

const (
	VmIdle int32 = iota
	VmStart
	VmRunning
	VmStoping
)

var vmStatusErr = errors.New("simBroker VM status error")

type simBroker int

func (b simBroker) Open(ch chan<- QuoteEvent) (Broker, error) {
	acctLock.Lock()
	defer acctLock.Unlock()
	var acct = account{evChan: ch}
	bb := simBroker(nAccounts)
	nAccounts++
	simAccounts = append(simAccounts, &acct)
	return bb, nil
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
		return vmStatusErr
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	maxAllocHeap = 0
	maxSysHeap = 0
	timeAtMaxAlloc = 0
	atomic.StoreInt32(&simStatus, VmStart)
	// load Bars
	// build ticks
	atomic.StoreInt32(&simStatus, VmRunning)
	return nil
}

func (b simBroker) Stop() error {
	switch atomic.LoadInt32(&simStatus) {
	case VmIdle, VmStoping:
		return nil
	case VmRunning:
	default:
		return vmStatusErr
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
		return vmStatusErr
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
	acct := simAccounts[int(b)]
	return acct.equity
}

func (b simBroker) Balance() float64 {
	acct := simAccounts[int(b)]
	return acct.balance
}

func (b simBroker) Cash() float64 {
	acct := simAccounts[int(b)]
	return acct.fund
}

func (b simBroker) FreeMargin() float64 {
	acct := simAccounts[int(b)]
	return acct.equity - acct.margin
}

func (b simBroker) SendOrder(sym string, dir OrderDirT, qty int, prc float64, stopL float64) int {
	simVmLock.Lock()
	defer simVmLock.Unlock()
	// tobe fix
	// verify, put to orderbook
	return 0
}

func (b simBroker) CancelOrder(oid int) {
	//acct := simAccounts[int(b)]
	if oid >= nOrders {
		return
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	// remove order from orderbook
}

func (b simBroker) CloseOrder(oId int) {
	//acct := simAccounts[int(b)]
	// if open, close with market
	// if stoploss, remove stoploss, change to market
	if oId >= nOrders {
		return
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	// if order open or partfill, changed to market order
}

func (b simBroker) GetOrder(oId int) *OrderType {
	if oId >= nOrders {
		return nil
	}
	orderLock.RLock()
	defer orderLock.RUnlock()
	return &simOrders[oId].OrderType
}

func (b simBroker) GetOrders() []int {
	acct := simAccounts[int(b)]
	return acct.orders
}

func (b simBroker) GetPosition(sym string) (vPos PositionType) {
	acct := simAccounts[int(b)]
	for _, v := range acct.pos {
		if v.Symbol == sym {
			vPos = v
			return
		}
	}
	return
}

func (b simBroker) GetPositions() []PositionType {
	acct := simAccounts[int(b)]
	return acct.pos
}

//go:noinline
func (b simBroker) TimeCurrent() DateTimeMs {
	return simCurrent
}

var simTrader simBroker

func init() {
	RegisterBroker("simBroker", simTrader)
}
