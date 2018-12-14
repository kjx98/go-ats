package ats

import "errors"

type Strategyer interface {
	Init(c Config) (Strategyer, error) // Initialize strategy paramter, universe in Config
	OnTick(sym string)                 // sym tick/quotes updated
	OnBar(sym string, period Period)   // bar with period updated
	DeInit()                           // Destroy interface/state
}

var stratExist = errors.New("Strategy registered")
var stratNotExist = errors.New("Strategy not registered")
var strats map[string]Strategyer = map[string]Strategyer{}

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
