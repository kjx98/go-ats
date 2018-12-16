package ats

import "errors"

// Parameter could be int/uint/float like type
type Parameter struct {
	Name  string
	Value interface{}
}

type Context struct {
	Broker
	Config
	GetBars func(sym string, period Period) (res *Bars, err error)
}

// universe of Strategyer should never intersection
// one symbol/ticker can only be processed by one Strategyer
// 		Context.Config items   "Universe":[]string, "Param":[]float64
type Strategyer interface {
	ParamSet() []Parameter               // Return parameter set for strategy
	Init(c *Context) (Strategyer, error) // Initialize strategy paramter, universe in c.Config
	OnTick(sym string)                   // sym tick/quotes updated
	OnBar(sym string, period Period)     // bar with period updated
	DeInit()                             // Destroy interface/state
}

var stratExist = errors.New("Strategy registered")
var stratNotExist = errors.New("Strategy not registered")
var strats map[string]Strategyer = map[string]Strategyer{}

func (c *Context) stratGetBars(sym string, period Period) (*Bars, error) {
	return getBars(sym, period, c.TimeCurrent())
}

func newContext(br Broker) *Context {
	var c = Context{Broker: br}
	c.GetBars = c.stratGetBars
	return &c
}

// RegisterStrategy should be called from init()
func RegisterStrategy(name string, inf Strategyer) error {
	if _, ok := strats[name]; ok {
		return stratExist
	}
	strats[name] = inf
	return nil
}

func loadStrategy(name string) (Strategyer, error) {
	if b, ok := strats[name]; ok {
		return b, nil
	}
	return nil, brokerNotExist
}