package ats

import (
	"errors"
)

// OrderDir current only Buy/Sell, no hedge position
type OrderDirT int32

const (
	OrderBuy   OrderDirT = 0
	OrderSell  OrderDirT = 1
	OrderCover OrderDirT = 2
	OrderClose OrderDirT = 3
)

// Order Sign
//	return 1   for Buy
//	return -1  for Sell
func (orDir OrderDirT) Sign() int {
	return (1 - 2*int(orDir&1))
}

// PosOffset
//	return true for position close/offset
func (orDir OrderDirT) PosOffset() bool {
	return orDir > OrderSell
}

func (orDir OrderDirT) String() string {
	switch orDir {
	case OrderBuy:
		return "Buy"
	case OrderSell:
		return "Sell"
	case OrderCover:
		return "Cover"
	case OrderClose:
		return "SellClose"
	}
	return "NA"
}

type OrderStatusT int32

const (
	OrderNil        OrderStatusT = 0
	OrderNew        OrderStatusT = 1
	OrderAccept     OrderStatusT = 2
	OrderPartFilled OrderStatusT = 3
	OrderFilled     OrderStatusT = 4
	OrderCanceled   OrderStatusT = 5
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
	Symbol    string
	Positions int
	PosFreeze int
	AvgPrice  float64 // average price for position
}

type Broker interface {
	Open() (Broker, error) // on success return interface pointer
	Login(uname string, key string) error
	SubscribeQuotes([]QuoteSubT) error

	GetEquity() float64                                                                      // return equity value current
	GetBalance() float64                                                                     // Balance after last settlement
	GetCash() float64                                                                        // available free cash
	GetFreeMargin() float64                                                                  // availble free margin
	OrderSend(sym string, dir OrderDirT, qty int, prc float64, stopL float64, magic int) int // return >=0 on success
	OrderCancel(oid int)                                                                     // Cancel Order
	GetOrder(oId int) *OrderType
	GetOrders() []OrderType
	GetPositions() []PositionType
	FlushQuotes()            // sync quotes from broker
	TimeCurrent() DateTimeMs // return current time of broker server in millisecond timestamp
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
