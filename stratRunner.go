package ats

import (
	"errors"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/kjx98/golib/ini"
)

type strategyRunner struct {
	evChan   chan QuoteEvent
	symStrat map[string]bool
	strats   map[string]Strategyer
	contxt   *Context
}

// buildParam ...	build params from ini config
func buildParam(c *ini.IniConfig, stratName string, param []Parameter) []float64 {
	pLen := len(param)
	if pLen == 0 {
		return []float64{}
	}
	res := make([]float64, pLen)
	for i, pp := range param {
		switch pp.Value.(type) {
		case int8, int16, int32, int64:
			pVal := c.GetConfigInt(stratName, pp.Name, int(reflect.ValueOf(pp.Value).Int()))
			res[i] = float64(pVal)
		case uint8, uint16, uint32, uint64:
			pVal := c.GetConfigInt(stratName, pp.Name, int(reflect.ValueOf(pp.Value).Uint()))
			res[i] = float64(pVal)
		case float32, float64:
			pVal := c.GetConfigDouble(stratName, pp.Name, reflect.ValueOf(pp.Value).Float())
			res[i] = pVal
		}
	}
	return res
}

// setParam ... set params from Config
func setParam(c Config, stratName string, param []Parameter) []float64 {
	pLen := len(param)
	if pLen == 0 {
		return []float64{}
	}
	res := make([]float64, pLen)
	for i, pp := range param {
		pName := stratName + "." + pp.Name
		switch pp.Value.(type) {
		case int8, int16, int32, int64:
			pVal := c.GetInt(pName, int(reflect.ValueOf(pp.Value).Int()))
			res[i] = float64(pVal)
		case uint8, uint16, uint32, uint64:
			pVal := c.GetInt(pName, int(reflect.ValueOf(pp.Value).Uint()))
			res[i] = float64(pVal)
		case float32, float64:
			pVal := c.GetFloat64(pName, reflect.ValueOf(pp.Value).Float())
			res[i] = pVal
		}
	}
	return res
}

func newStrategyRunner() *strategyRunner {
	var res strategyRunner
	res.evChan = make(chan QuoteEvent, 10)
	res.symStrat = map[string]bool{}
	res.strats = map[string]Strategyer{}
	return &res
}

// SetStrategyParam ...	set Strategy params using Config
//				values of config item format StratName.ParamName
func (sc *strategyRunner) SetStrategyParam(c Config) error {
	for stName, b := range sc.strats {
		params := setParam(c, stName, b.ParamSet())
		sc.contxt.Put("Param", params)
		if ss, err := b.Init(sc.contxt); err == nil {
			sc.strats[stName] = ss
		} else {
			return err
		}
	}
	return nil
}

func (sc *strategyRunner) loadStrategy(fname string) (err error) {
	if len(sc.strats) != 0 {
		// already load strategy
		return
	}
	cf, err := ini.ParserConfig(fname, false)
	if err != nil {
		return
	}
	bName := cf.GetConfig("Config", "Broker", "")
	br, err := openBroker(bName, sc.evChan)
	if err != nil {
		return
	}
	sc.contxt = newContext(br)
	var autoNew bool
	if cf.GetConfigInt("Config", "NewSymbolInfo", 0) != 0 {
		autoNew = true
	}
	if cf.GetConfigInt("Config", "RunTick", 0) != 0 {
		sc.contxt.Put("RunTick", 1)
	}
	stratsN := strings.Split(cf.GetConfig("Config", "Strategy", ""), ",")
	for _, stName := range stratsN {
		if b, ok := stratsMap[stName]; ok {
			universe := strings.Split(cf.GetConfig(stName, "Universe", ""), ",")
			sc.contxt.Put("Universe", universe)
			// build param for Strategy
			params := buildParam(cf, stName, b.ParamSet())
			sc.contxt.Put("Param", params)
			if ss, err := b.Init(sc.contxt); err == nil {
				// process universe
				sc.strats[stName] = ss
				universe = sc.contxt.GetStrings("Universe")
				for _, sym := range universe {
					if _, err := GetSymbolInfo(sym); err != nil {
						if autoNew {
							newSymbolInfo(sym)
							if _, err := GetSymbolInfo(sym); err != nil {
								// can't buildSymbolInfo, skip
								continue
							}
						} else {
							continue
						}
					}
					if _, ok := sc.symStrat[sym]; ok {
						// already
						continue
					}
					sc.symStrat[sym] = true
				}
			}
		}
	}
	if len(sc.strats) == 0 {
		err = errNoActiveStrategy
		return
	}
	// subscribe quotes
	subs := []QuoteSubT{}
	for sym := range sc.symStrat {
		si, err := GetSymbolInfo(sym)
		if err != nil {
			continue
		}
		var subo = QuoteSubT{Symbol: sym}
		subo.QuotesPtr = si.getQuotesPtr()
		subs = append(subs, subo)
	}
	if err = br.SubscribeQuotes(subs); err != nil {
		log.Error("Broker SubscribeQuotes", err)
		return
	}
	return
}

var errNoEventChannel = errors.New("No Event Channel")
var errNoStrategy = errors.New("No Strategy loaded")
var errNoActiveStrategy = errors.New("No active Strategy")

func (sc *strategyRunner) emitEvent(si *SymbolInfo, evID int) {
	for _, strat := range sc.strats {
		switch Period(evID) {
		case 0:
			strat.OnTick(si.Ticker)
		case Min1, Min5, Hour1, Daily:
			strat.OnBar(si.Ticker, Period(evID))
		}
	}
}

var wg sync.WaitGroup

func (sc *strategyRunner) runStrategy() error {
	if sc.evChan == nil {
		return errNoEventChannel
	}
	if len(sc.strats) == 0 {
		return errNoStrategy
	}
	sc.contxt.Broker.Start(sc.contxt.Config)
	wg.Add(1)
	// subscribe quotes
	go func() {
		// process event
		defer wg.Done()
		for {
			select {
			case ev, ok := <-sc.evChan:
				if !ok {
					return
				}
				if ev.EventID < 0 {
					// run out of sample Bars
					return
				}
				// process ev
				if si, err := GetSymbolInfo(ev.Symbol); err == nil {
					sc.emitEvent(&si, ev.EventID)
				}
			}
			runtime.Gosched()
		}
		// never reach
	}()
	return nil
}

func (sc *strategyRunner) stopStrategy() {
	//for multiple running, never close evChan
	//close(sc.evChan)
	wg.Wait()
	for _, ss := range sc.strats {
		ss.DeInit()
	}
}
