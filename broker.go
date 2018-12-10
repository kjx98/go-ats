package ats

import (
	"errors"
)

// OrderDir current only Buy/Sell, no hedge position
type OrderDirT int32

const (
	OrderBuy  OrderDirT = 0
	OrderSell OrderDirT = 1
)

type OrderStatusT int32

const (
	OrderNil        OrderStatusT = 0
	OrderNew        OrderStatusT = 1
	OrderAccept     OrderStatusT = 2
	OrderPartFilled OrderStatusT = 3
	OrderFilled     OrderStatusT = 4
	OrderCanceled   OrderStatusT = 5
)

// Order struct
type OrderType struct {
	Symbol      string
	Dir         OrderDirT
	Prc         float64
	Qty         int
	StopPrice   float64
	ProfitPrice float64
	QtyFilled   int
	Status      OrderStatusT
}

type PositionType struct {
	Symbol    string
	Positions int
	PosFreeze int
	AvgPrice  float64 // average price for position
}

type Broker interface {
	Open() (Broker, error) // on success return interface pointer
	Login(uname string, key string) error
	SubscribeQuotes([]QuoteSubT) error
	GetEquity() float64  // return equity value current
	GetBalance() float64 // Balance after last settlement
	GetCash() float64    // available free cash
	GetOrder(oId int) *OrderType
	GetOrders() []OrderType
	GetPositions() []PositionType
	FlushQuotes() // sync quotes from broker
}

var brokerExist = errors.New("Broker registered")
var brokerNotExist = errors.New("Borker not registered")
var brokers map[string]Broker = map[string]Broker{}

func RegisterBroker(name string, inf Broker) error {
	if _, ok := brokers[name]; ok {
		return brokerExist
	}
	brokers[name] = inf
	return nil
}

func OpenBroker(name string) (Broker, error) {
	if b, ok := brokers[name]; ok {
		return b.Open()
	}
	return nil, brokerNotExist
}
