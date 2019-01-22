package ats

import (
	"errors"
)

// OrderDirT current only Buy/Sell, no hedge position
type OrderDirT int32

// OrderDirBuy ...	Buy long open
// OrderDirSell ...	Sell short open
// OrderDirCover ...	Buy cover short
// OrderDirClose ...	Sell close long
const (
	OrderDirBuy OrderDirT = iota
	OrderDirSell
	OrderDirCover
	OrderDirClose
)

// Sign ...	for order dir
//	return 1   for Buy
//	return -1  for Sell
func (orDir OrderDirT) Sign() int {
	return (1 - 2*int(orDir&1))
}

// IsOffset ...	return true for position close/offset
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

// OrderStatusT ... order status
type OrderStatusT int32

// OrderNil ...	nil order
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

// OrderType ... struct
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

// PositionType ...		position for symbol of account
type PositionType struct {
	fKey      SymbolKey
	Positions int
	PosFreeze int
	AvgPrice  float64 // average price for position
}

// Broker ...	interface for abstract broker
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
	CancelOrder(oID int) error                                                    // Cancel Order
	CloseOrder(oID int)
	GetOrder(oID int) *OrderType
	GetOrders() []int
	GetPosition(sym string) PositionType
	GetPositions() []PositionType
	TimeCurrent() DateTimeMs // return current time of broker server in millisecond timestamp
}

var errBrokerExist = errors.New("Broker registered")
var errBrokerNotExist = errors.New("Borker not registered")
var brokers = map[string]Broker{}
var defaultBroker = "simBroker"

// RegisterBroker ... register broker with name
func RegisterBroker(name string, inf Broker) error {
	if _, ok := brokers[name]; ok {
		return errBrokerExist
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
	return nil, errBrokerNotExist
}
