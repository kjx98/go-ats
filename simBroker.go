package ats

import (
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

func (b simBroker) Start(c Config) error {
	// read Config, ...
	// start goroutine for simulate/backtesting
	return nil
}

func (b simBroker) Stop() error {
	return nil
}

func (b simBroker) SubscribeQuotes([]QuoteSubT) error {
	// prepare Bars
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
	// tobe fix
	return 0
}

func (b simBroker) CancelOrder(oid int) {
	acct := simAccounts[int(b)]
	if oid >= len(acct.orders) {
		return
	}
}

func (b simBroker) CloseOrder(oId int) {
	acct := simAccounts[int(b)]
	// if open, close with market
	// if stoploss, remove stoploss, change to market
	if oId >= len(acct.orders) {
		return
	}
}

func (b simBroker) GetOrder(oId int) *OrderType {
	orderLock.RLock()
	defer orderLock.RUnlock()
	if oId >= nOrders {
		return nil
	}

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

func (b simBroker) TimeCurrent() DateTimeMs {
	return simCurrent
}

var simTrader simBroker

func init() {
	RegisterBroker("simBroker", simTrader)
}
