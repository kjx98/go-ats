package ats

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/op/go-logging"
	//yaml "gopkg.in/yaml.v2"
)

const (
	maxInstruments int = 4096
)

// Market  Exchange/combo national markets
// VolMin  minimal volume of order
// VolMax  maximal volume of single order
// VolStep volume step of order, volume tick
// PriceStep	minimal price step
// PriceDigits
// VolDigits
// Margin	suppose initial margin and maintain margin are same, no support for options
// IsForex	Forex/CFD ... OTC instrument without last/sales
// CommissionType		0	per Amount, 1 Per Lot, 2 Per Trade
type symbolBase struct {
	Market         string  `json:"market,omitempty"`
	VolMin         int     `json:"volumeMin"`
	VolMax         int     `json:"volumeMax"`
	VolStep        int     `json:"volumeStep"`
	PriceStep      float64 `json:"priceStep"`
	PriceDigits    int     `json:"digits,omitempty"`
	VolDigits      int     `json:"volumeDigits,omitempty"`
	LotSize        int     `json:"lotSize,omitempty"`
	Margin         float64 `json:"margin,omitempty"`
	IsForex        bool    `json:"forex,omitempty"`
	DefSpread      int32   `json:"defSpread,omitempty"`
	CurrencySym    string  `json:"currency,omitempty"`
	CommissionType int     `json:"commisssionType,omitempty"`
	CommissionRate float64 `json:"commissionRate,omitempty"`
	bMargin        bool
}

// SymbolKey ... fast key for symbol, based from 1
type SymbolKey int

// SymbolInfo symbol traits of instrument
// fKey link Bars/DayTA/MinTA etc, index from 1 .. count
type SymbolInfo struct {
	Ticker string
	*symbolBase
	deliverMonth int
	fKey         SymbolKey
	Upper        float64
	Lower        float64
	quote        Quotes
}

// FastKey for internal
func (s *SymbolInfo) FastKey() SymbolKey {
	return s.fKey
}

func (s *SymbolInfo) Digits() int {
	return s.PriceDigits
}

func (s *SymbolInfo) VolumeDigits() int {
	return s.VolDigits
}

func (s *SymbolInfo) Multi() float64 {
	return digitMulti(s.PriceDigits)
}

func (s *SymbolInfo) Divi() float64 {
	return digitDiv(s.PriceDigits)
}

// PriceNormal normal Price for order
func (s *SymbolInfo) PriceNormal(p float64) float64 {
	p = math.Floor(p/s.PriceStep) * s.PriceStep
	if p < s.Lower {
		p = s.Lower
	} else if s.Upper > 0 && p > s.Upper {
		p = s.Upper
	}
	return p
}

// GetQuotes return quotes for symbol
func (s *SymbolInfo) GetQuotes() Quotes {
	return s.quote
}

// return ref for quotes of symbol, used by broker quotes feed
func (s *SymbolInfo) getQuotesPtr() *Quotes {
	return &s.quote
}

// CalcProfit calc order profit according to price and volume position
//		volume < 0 for short
func (s *SymbolInfo) CalcProfit(openP, closeP float64, volume int) float64 {
	fVol := float64(volume)
	if s.LotSize > 0 {
		fVol *= float64(s.LotSize)
	}
	if vd := s.VolDigits; vd > 0 {
		mm := digitDiv(vd)
		fVol *= mm
	}
	res := (closeP - openP) * fVol
	return res
}

// CalcVolume calc order quantity according to price and value amount in full margin
func (s *SymbolInfo) CalcVolume(amt float64, p float64) float64 {
	if s.LotSize > 0 {
		p *= float64(s.LotSize)
	}
	res := amt / p
	if s.bMargin {
		res /= s.Margin
	}
	volStep := float64(s.VolStep)
	volMin := float64(s.VolMin)
	volMax := float64(s.VolMax)
	if vd := s.VolDigits; vd > 0 {
		mm := digitDiv(vd)
		volStep *= mm
		volMin *= mm
		volMax *= mm
	}
	res = math.Floor(res/volStep) * volStep
	if res < volMin {
		res = volMin
	} else if res > volMax {
		res = volMax
	}
	return res
}

// CalcRiskVolume ... calc order quantity according to price and value amount according to margin
func (s *SymbolInfo) CalcRiskVolume(amt float64, riskPrice float64) float64 {
	if s.LotSize > 0 {
		riskPrice *= float64(s.LotSize)
	}
	res := amt / riskPrice
	volStep := float64(s.VolStep)
	volMin := float64(s.VolMin)
	volMax := float64(s.VolMax)
	if vd := s.VolDigits; vd > 0 {
		mm := digitDiv(vd)
		volStep *= mm
		volMin *= mm
		volMax *= mm
	}
	res = math.Floor(res/volStep) * volStep
	if res < volMin {
		res = volMin
	} else if res > volMax {
		res = volMax
	}
	return res
}

func (s *SymbolInfo) String() (res string) {
	margin := ""
	if s.bMargin {
		margin = fmt.Sprintf("margin(%.2f)", s.Margin)
	}
	if vd := s.VolDigits; vd > 0 {
		mm := digitDiv(vd)
		volMin := float64(s.VolMin) * mm
		volMax := float64(s.VolMax) * mm
		volStep := float64(s.VolStep) * mm
		res = fmt.Sprintf("%s@%s fKey(%d) Vol(%.*f/%.*f)(%.*f) PrcStep(%.*f) %s forex(%v)",
			s.Ticker, s.Market, s.fKey, vd, volMin, vd, volMax, vd, volStep,
			s.PriceDigits, s.PriceStep, margin, s.IsForex)
	} else {
		res = fmt.Sprintf("%s@%s fkey(%d) Vol(%d/%d)(%d) PrcStep(%.*f) %s forex(%v)",
			s.Ticker, s.Market, s.fKey, s.VolMin, s.VolMax, s.VolStep,
			s.PriceDigits, s.PriceStep, margin, s.IsForex)
	}
	return
}

// symbol Template used for autoNew tickers
type symbolTemplate struct {
	TickerPrefix string `json:"ticker"`
	Name         string
	Base         symbolBase
	TickerLen    int  `json:"tickerLen,omitempty"`
	DateLen      int  `json:"dateLen,omitempty"`
	USticker     bool `json:"usTicker,omitempty"`
	Bregexp      bool `json:"regexp,omitempty"`
	exp          *regexp.Regexp
}

// TickStart, TickEnd only
// Daily Data should used whole
type tickConf struct {
	symbol    string
	TickStart uint32 `json:"start,omitempty"`
	TickEnd   uint32 `json:"end,omitempty"`
}

var fDiv = [...]float64{100.0, 10.0, 1.0, 0.1, 0.01, 0.001, 0.0001,
	0.00001, 0.000001}
var fMulti = [...]float64{0.01, 0.1, 1.0, 10.0, 100.0, 1000.0, 10000.0,
	100000.0, 1000000.0}

func digitMulti(ndigit int) float64 {
	if ndigit < -2 || ndigit > 6 {
		return 1.0
	}
	return fMulti[ndigit+2]
}

func digitDiv(ndigit int) float64 {
	if ndigit < -2 || ndigit > 6 {
		return 1.0
	}
	return fDiv[ndigit+2]
}

func DigitToInt(v float64, ndigit int) int {
	res := v * digitMulti(ndigit)
	return int(res)
}

func DigitFromInt(v int, ndigit int) float64 {
	res := float64(v) * digitDiv(ndigit)
	return res
}

func (t *symbolTemplate) String() (res string) {
	margin := ""
	if t.Base.bMargin {
		margin = fmt.Sprintf("%f", t.Base.Margin)
	}
	if t.Bregexp {
		margin += " regexp"
	}
	if vd := t.Base.VolDigits; vd > 0 {
		mm := digitDiv(vd)
		volMin := float64(t.Base.VolMin) * mm
		volMax := float64(t.Base.VolMax) * mm
		volStep := float64(t.Base.VolStep) * mm
		res = fmt.Sprintf("%s@%s Vol(%.*f/%.*f)(%.*f) PrcStep(%d,%.*f) %s %d %d",
			t.TickerPrefix, t.Base.Market, vd, volMin, vd, volMax, vd, volStep,
			t.Base.PriceDigits, t.Base.PriceDigits, t.Base.PriceStep, margin,
			t.TickerLen, t.DateLen)
	} else {
		res = fmt.Sprintf("%s@%s Vol(%d/%d)(%d) PrcStep(%d,%.*f) %s %d %d",
			t.TickerPrefix, t.Base.Market, t.Base.VolMin, t.Base.VolMax,
			t.Base.VolStep, t.Base.PriceDigits, t.Base.PriceDigits,
			t.Base.PriceStep, margin, t.TickerLen, t.DateLen)
	}
	return
}

type symbolTemps []symbolTemplate

func (data symbolTemps) Len() int {
	return len(data)
}

// prefer longer match
func (data symbolTemps) Less(i, j int) bool {
	if data[i].Bregexp && data[j].Bregexp {
		return len(data[i].TickerPrefix) > len(data[j].TickerPrefix)
	} else if !data[i].Bregexp && !data[j].Bregexp {
		return len(data[i].TickerPrefix) > len(data[j].TickerPrefix)
	} else if data[i].Bregexp {
		return false
	}
	return true

}

func (data symbolTemps) Swap(i, j int) {
	data[i], data[j] = data[j], data[i]
}

var log = logging.MustGetLogger("go-ats")
var initTemp = symbolTemps{
	{TickerPrefix: "cu",
		Name: "copper future contract",
		Base: symbolBase{Market: "SHFE",
			VolMin:      1,
			VolMax:      2000,
			VolStep:     1,
			PriceStep:   100,
			PriceDigits: 0,
			VolDigits:   0},
		TickerLen: 6,
		DateLen:   4,
		USticker:  false,
	},
	{TickerPrefix: "sh6",
		Name: "Shanghai A stock",
		Base: symbolBase{Market: "SHSE",
			VolMin:      100,
			VolMax:      1000000,
			VolStep:     100,
			PriceStep:   0.01,
			PriceDigits: 2,
			VolDigits:   0},
		TickerLen: 8,
		DateLen:   0,
		USticker:  false,
	},
	{TickerPrefix: "sh5",
		Name: "Shanghai ETF",
		Base: symbolBase{Market: "SHSE",
			VolMin:      100,
			VolMax:      1000000,
			VolStep:     100,
			PriceStep:   0.001,
			PriceDigits: 3,
			VolDigits:   0},
		TickerLen: 8,
		DateLen:   0,
		USticker:  false,
	},
	{TickerPrefix: "sh204",
		Name: "Shanghai Repo",
		Base: symbolBase{Market: "SHSE",
			VolMin:      10,
			VolMax:      100000,
			VolStep:     10,
			PriceStep:   0.001,
			PriceDigits: 3,
			VolDigits:   0},
		TickerLen: 8,
		DateLen:   0,
		USticker:  false,
	},
}

var symInfos = map[string]*SymbolInfo{}
var symInfoCaches = []SymbolInfo{}
var usDeliverMonth = " FGHJKMNQUVXZ"
var errNoSuchSymbol = errors.New("no such symbol")
var initTicks = map[string]tickConf{}

func GetSymbolInfo(sym string) (SymbolInfo, error) {
	if res, ok := symInfos[sym]; ok {
		return *res, nil
	}
	return SymbolInfo{}, errNoSuchSymbol
}

func (fkey SymbolKey) SymbolInfo() (*SymbolInfo, error) {
	idx := int(fkey)
	if idx <= 0 || idx > len(symInfoCaches) {
		return nil, errNoSuchSymbol
	}
	return &symInfoCaches[idx-1], nil
}

var nInstruments int
var instRWlock sync.RWMutex
var jpySymbolBasePtr *symbolBase

func newSymbolInfo(sym string) {
	sLen := len(sym)
	if sLen == 0 {
		return
	}
	instRWlock.RLock()
	if _, ok := symInfos[sym]; ok {
		instRWlock.RUnlock()
		return
	}
	// maxium instruments reach, no more instrument add
	if nInstruments == maxInstruments {
		return
	}
	instRWlock.RUnlock()
	var symInfo = SymbolInfo{}
	for i := 0; i < len(initTemp); i++ {
		if sLen > initTemp[i].TickerLen {
			continue
		}
		if initTemp[i].Bregexp {
			if initTemp[i].exp == nil {
				continue
			}
			if sm := initTemp[i].exp.FindString(sym); sm == "" {
				continue
			}
		} else if len(initTemp[i].TickerPrefix) > 0 {
			if initTemp[i].TickerLen != sLen {
				continue
			}
			if sym[:len(initTemp[i].TickerPrefix)] != initTemp[i].TickerPrefix {
				continue
			}
		}
		switch initTemp[i].DateLen {
		case 2:
			if sLen < 3 || !initTemp[i].USticker {
				continue
			}
			if sym[sLen-1] < '0' || sym[sLen-1] > '9' {
				continue
			}
			mon := strings.IndexByte(usDeliverMonth, sym[sLen-2])
			if mon < 0 {
				// deliver Month no exist
				return
			}
			symInfo.deliverMonth = mon
		case 3:
			if sLen < 4 {
				continue
			}
			if res, err := strconv.Atoi(sym[sLen-2:]); err != nil {
				continue
			} else if res < 1 || res > 12 {
				return
			} else {
				symInfo.deliverMonth = res
			}
		case 4:
			if sLen < 5 {
				continue
			}
			if res, err := strconv.Atoi(sym[sLen-2:]); err != nil {
				continue
			} else if res < 1 || res > 12 {
				return
			} else {
				symInfo.deliverMonth = res
			}
		default:
			continue
		case 0:
		}
		symInfo.Ticker = sym
		symInfo.symbolBase = &initTemp[i].Base
		if symInfo.IsForex && sym[3:] == "JPY" {
			if jpySymbolBasePtr == nil {
				jpyBase := initTemp[i].Base
				jpySymbolBasePtr = &jpyBase
				jpyBase.PriceDigits = 3
				jpyBase.PriceStep = 0.001
			}
			symInfo.symbolBase = jpySymbolBasePtr
		}
		instRWlock.Lock()
		defer instRWlock.Unlock()

		symIdx := nInstruments
		nInstruments++
		symInfo.fKey = SymbolKey(nInstruments)

		symInfoCaches = append(symInfoCaches, symInfo)
		symInfos[sym] = &symInfoCaches[symIdx]
		//symInfos[sym] = &symInfo
		return
	}
}

var symbolTempOnce sync.Once

func initSymbols() {
	symbolTempOnce.Do(func() {
		if bb, err := ioutil.ReadFile("symbols.json"); err == nil {
			var symTemps map[string]symbolTemplate
			if err := json.Unmarshal(bb, &symTemps); err == nil {
				initTemp = []symbolTemplate{}
				for _, ss := range symTemps {
					if ss.Bregexp {
						if ss.exp, err = regexp.Compile(ss.TickerPrefix); err != nil {
							continue
						}
					}
					initTemp = append(initTemp, ss)
				}
			} else {
				log.Warning("Decode symbols.json", err)
			}
		}
		sort.Sort(initTemp)
		// validate initTemp
		for i := 0; i < len(initTemp); i++ {
			if initTemp[i].Base.Margin <= 0 || initTemp[i].Base.Margin == 1.0 {
				initTemp[i].Base.bMargin = false
			} else {
				initTemp[i].Base.bMargin = true
			}
			if initTemp[i].Base.VolStep <= 0 {
				initTemp[i].Base.VolStep = 1
			}
			if initTemp[i].Base.PriceStep <= 0 {
				initTemp[i].Base.PriceStep = 1
			}
			// digitMulti/digitDiv validate price/volume digits
			/*
				if initTemp[i].Base.VolDigits > len(fMulti)-3 {
					initTemp[i].Base.VolDigits = len(fMulti) - 3
				}
				if initTemp[i].Base.PriceDigits > len(fMulti)-3 {
					initTemp[i].Base.PriceDigits = len(fMulti) - 3
				}
			*/
		}
		if bb, err := ioutil.ReadFile("ticks.json"); err == nil {
			var symMap map[string]tickConf
			if err := json.Unmarshal(bb, &symMap); err == nil {
				for sym, ss := range symMap {
					newSymbolInfo(sym)
					//log.Info("Load Ticker", sym)
					if _, err := GetSymbolInfo(sym); err == nil {
						var ti = tickConf{sym, ss.TickStart, ss.TickEnd}
						initTicks[sym] = ti
					}
				}
			} else {
				log.Error("Decode ticks.json", err)
			}
		}
	})
}

//	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`
func init() {
	var format = logging.MustStringFormatter(
		`%{color}%{time:01-02 15:04:05}  ▶ %{level:.4s} %{color:reset} %{message}`,
	)

	logback := logging.NewLogBackend(os.Stderr, "", 0)
	logfmt := logging.NewBackendFormatter(logback, format)
	logging.SetBackend(logfmt)
	initSymbols()
}
