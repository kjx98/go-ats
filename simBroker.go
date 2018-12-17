package ats

import (
	"errors"
	"sync"
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
	OrderType
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
var simStatus int

const (
	VmIdle int = iota
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
	switch simStatus {
	case VmIdle:
	case VmStart, VmRunning:
		return nil
	default:
		return vmStatusErr
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	simStatus = VmStart
	// load Bars
	// build ticks
	simStatus = VmRunning
	return nil
}

func (b simBroker) Stop() error {
	switch simStatus {
	case VmIdle, VmStoping:
		return nil
	case VmRunning:
	default:
		return vmStatusErr
	}
	simVmLock.Lock()
	defer simVmLock.Unlock()
	simStatus = VmStoping
	// stop Bar feed
	simStatus = VmIdle
	return nil
}

func (b simBroker) SubscribeQuotes([]QuoteSubT) error {
	if simStatus != VmIdle {
		return vmStatusErr
	}
	// prepare Bars
	// maybe Once load?
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
