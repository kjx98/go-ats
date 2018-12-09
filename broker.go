package ats

import (
)


// OrderDir current only Buy/Sell, no hedge position
type OrderDirT	int32
const (
	OrderBuy OrderDirT = 0
	OrderSell OrderDirT = 1
)

type OrderStatusT int32
const (
	OrderNil OrderStatusT= 0
	OrderNew	OrderStatusT= 1
	OrderAccept	OrderStatusT= 2
	OrderPartFilled OrderStatusT= 3
	OrderFilled OrderStatusT= 4
	OrderCanceled OrderStatusT= 5
)


// Order struct
type OrderType struct {
	Symbol		string
	Dir			OrderDirT
	Prc			float64
	Qty			int
	StopPrice	float64
	ProfitPrice	float64
	QtyFilled	int
	Status		OrderStatusT
}

type PositionType struct {
	Symbol		string
	LongPos		int
	ShortPos	int
	LongFreeze	int
	ShortFreeze	int
	LongPrice	float64	// average price for long position
	ShortPrice	float64	// average price for short position
}

type Broker interface {
	Open()		(*Broker,error)	// on success return interface pointer
	Login(uname string, key string) error
	GetEquity()	float64				// return equity value current
	GetBalance()	float64	// Balance after last settlement
	GetCash()	float64		// available free cash
	GetOrder(oId int) *OrderType
	GetOrders()	[]OrderType
	GetPositions() []PositionType
	FlushQuotes()		// sync quotes from broker
}