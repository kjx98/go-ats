package ats

type simBroker struct {
	startTime  timeT64
	endTime    timeT64
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
	current    DateTimeMs
	evChan     chan<- QuoteEvent
	orders     []OrderType
	pos        []PositionType
}

func (b *simBroker) Open(ch chan<- QuoteEvent) (Broker, error) {
	var bb = simBroker{evChan: ch}
	return &bb, nil
}

func (b *simBroker) Start(c Config) error {
	// read Config, ...
	// start goroutine for simulate/backtesting
	return nil
}

func (b *simBroker) Stop() error {
	return nil
}

func (b *simBroker) SubscribeQuotes([]QuoteSubT) error {
	// prepare Bars
	return nil
}

func (b *simBroker) GetEquity() float64 {
	return b.equity
}

func (b *simBroker) GetBalance() float64 {
	return b.balance
}

func (b *simBroker) GetCash() float64 {
	return b.fund
}

func (b *simBroker) GetFreeMargin() float64 {
	return b.equity - b.margin
}

func (b *simBroker) SendOrder(sym string, dir OrderDirT, qty int, prc float64, stopL float64) int {
	// tobe fix
	return 0
}

func (b *simBroker) CancelOrder(oid int) {
	// find order, cancel
	if oid >= len(b.orders) {
		return
	}
}

func (b *simBroker) CloseOrder(oId int) {
	// if open, close with market
	// if stoploss, remove stoploss, change to market
	if oId >= len(b.orders) {
		return
	}
}

func (b *simBroker) GetOrder(oId int) *OrderType {
	if oId >= len(b.orders) {
		return nil
	}
	return &b.orders[oId]
}

func (b *simBroker) GetOrders() []OrderType {
	return b.orders
}

func (b *simBroker) GetPosition(sym string) (vPos PositionType) {
	for _, v := range b.pos {
		if v.Symbol == sym {
			vPos = v
			return
		}
	}
	return
}

func (b *simBroker) GetPositions() []PositionType {
	return b.pos
}

func (b *simBroker) TimeCurrent() DateTimeMs {
	return b.current
}

var simTrader = simBroker{}

func init() {
	RegisterBroker("simBroker", &simTrader)
}
