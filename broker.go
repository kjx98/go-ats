package ats

import (
	"errors"
)

// OrderDir current only Buy/Sell, no hedge position
type OrderDirT int32

const (
	OrderDirBuy OrderDirT = iota
	OrderDirSell
	OrderDirCover
	OrderDirClose
)

// Order Sign
//	return 1   for Buy
//	return -1  for Sell
func (orDir OrderDirT) Sign() int {
	return (1 - 2*int(orDir&1))
}

// PosOffset
//	return true for position close/offset
func (orDir OrderDirT) IsOffset() bool {
	return orDir > OrderDirSell
}

func (orDir OrderDirT) String() string {
	switch orDir {
	case OrderDirBuy:
		return "Buy"
	case OrderDirSell:
		return "Sell"
	case OrderDirCover:
		return "Cover"
	case OrderDirClose:
		return "SellClose"
	}
	return "NA"
}

type OrderStatusT int32

const (
	OrderNil OrderStatusT = iota
	OrderNew
	OrderAccept
	OrderPartFilled
	OrderFilled
	OrderCanceled
)

func (oSt OrderStatusT) String() string {
	switch oSt {
	case OrderNil:
		return "Nil"
	case OrderNew:
		return "New"
	case OrderAccept:
		return "Accept"
	case OrderPartFilled:
		return "PartFill"
	case OrderFilled:
		return "Filled"
	case OrderCanceled:
		return "Canceled"
	}
	return "Invalid"
}

// Order struct
//	Symbol  order symbol
//	Price	order price(or profit price if Stop not zero)
//	StopPrice	StopLoss price
//	Dir		OrderBuy, OrderSell
//	Qty		order quantity
//	QtyFilled
type OrderType struct {
	Symbol    string
	Price     float64
	StopPrice float64
	Dir       OrderDirT
	Qty       int
	QtyFilled int
	Magic     int
	Status    OrderStatusT
	OpenTime  DateTimeMs
	CloseTime DateTimeMs
	AvgPrice  float64
}

type PositionType struct {
	fKey      SymbolKey
	Positions int
	PosFreeze int
	AvgPrice  float64 // average price for position
}

type Broker interface {
	Open(ch chan<- QuoteEvent) (Broker, error) // on success return interface pointer
	Start(c Config) error                      // user/pwd, startDate/endDate ...
	Stop() error                               // stop broker, logout, cleanup
	SubscribeQuotes([]QuoteSubT) error

	Equity() float64                                                              // return equity value current
	Balance() float64                                                             // Balance after last settlement
	Cash() float64                                                                // available free cash
	FreeMargin() float64                                                          // availble free margin
	SendOrder(sym string, dir OrderDirT, qty int, prc float64, stopL float64) int // return oId >=0 on success
	CancelOrder(oid int) error                                                    // Cancel Order
	CloseOrder(oId int)
	GetOrder(oId int) *OrderType
	GetOrders() []int
	GetPosition(sym string) PositionType
	GetPositions() []PositionType
	TimeCurrent() DateTimeMs // return current time of broker server in millisecond timestamp
}

var brokerExist = errors.New("Broker registered")
var brokerNotExist = errors.New("Borker not registered")
var brokers map[string]Broker = map[string]Broker{}
var defaultBroker = "simBroker"

func RegisterBroker(name string, inf Broker) error {
	if _, ok := brokers[name]; ok {
		return brokerExist
	}
	brokers[name] = inf
	return nil
}

func openBroker(name string, ch chan<- QuoteEvent) (Broker, error) {
	if name == "" {
		name = defaultBroker
	}
	if b, ok := brokers[name]; ok {
		return b.Open(ch)
	}
	return nil, brokerNotExist
}
