package ats

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Market  Exchange/combo national markets
// VolMin  minimal volume of order
// VolMax  maximal volume of single order
// VolStep volume step of order, volume tick
// PriceStep	minimal price step
// PriceDigits
// VolDigits
// Margin	suppose initial margin and maintain margin are same, no support for options
type symbolBase struct {
	Market      string  `yaml:"market,omitempty"`
	VolMin      int     `yaml:"volumeMin"`
	VolMax      int     `yaml:"volumeMax"`
	VolStep     int     `yaml:"volumeStep"`
	PriceStep   float64 `yaml:"priceStep"`
	PriceDigits int     `yaml:"digits,omitempty"`
	VolDigits   int     `yaml:"volumeDigits,omitempty"`
	LotSize     int     `yaml:"lotSize,omitempty"`
	Margin      float64 `yaml:"margin,omitempty"`
	bMargin     bool
}

type SymbolInfo struct {
	Ticker string
	*symbolBase
	deliverMonth int
	Upper        float64
	Lower        float64
	Quotes
}

func (s *SymbolInfo) Digits() int {
	return s.PriceDigits
}

func (s *SymbolInfo) VolumeDigits() int {
	return s.VolDigits
}

func (s *SymbolInfo) PriceNormal(p float64) float64 {
	p = math.Floor(p/s.PriceStep) * s.PriceStep
	if p < s.Lower {
		p = s.Lower
	} else if s.Upper > 0 && p > s.Upper {
		p = s.Upper
	}
	return p
}

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
		volStep *= digitMulti(vd)
		volMin *= digitMulti(vd)
		volMax *= digitMulti(vd)
	}
	res = math.Floor(res/volStep) * volStep
	if res < volMin {
		res = volMin
	} else if res > volMax {
		res = volMax
	}
	return res
}

type symbolTemplate struct {
	TickerPrefix string `yaml:"ticker"`
	Name         string
	Base         symbolBase
	TickerLen    int  `yaml:"tickerLen,omitempty"`
	DateLen      int  `yaml:"dateLen,omitempty"`
	USticker     bool `yaml:"usTicker,omitempty"`
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

func (t symbolTemplate) String() (res string) {
	margin := ""
	if t.Base.Margin > 0 {
		margin = fmt.Sprintf("%f", t.Base.Margin)
	}
	if vd := t.Base.VolDigits; vd > 0 {
		mm := digitMulti(vd)
		volMin := float64(t.Base.VolMin) * mm
		volMax := float64(t.Base.VolMax) * mm
		volStep := float64(t.Base.VolStep) * mm
		res = fmt.Sprintf("%s@%s Vol(%.*f/%.*f)(%.*f) PrcStep(%.*f) %s %d %d,\n",
			t.TickerPrefix, t.Base.Market, vd, volMin, vd, volMax,
			vd, volStep, t.Base.PriceDigits, t.Base.PriceStep,
			margin, t.TickerLen, t.DateLen)
	} else {
		res = fmt.Sprintf("%s@%s Vol(%d/%d)(%d) PrcStep(%.*f) %s %d %d,\n",
			t.TickerPrefix, t.Base.Market, t.Base.VolMin, t.Base.VolMax,
			t.Base.VolStep, t.Base.PriceDigits, t.Base.PriceStep,
			margin, t.TickerLen, t.DateLen)
	}
	return
}

type symbolTemps []symbolTemplate

func (data symbolTemps) Len() int {
	return len(data)
}

func (data symbolTemps) Less(i, j int) bool {
	return len(data[i].TickerPrefix) < len(data[j].TickerPrefix)
}

func (data symbolTemps) Swap(i, j int) {
	data[i], data[j] = data[j], data[i]
}

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
var usDeliverMonth = " FGHJKMNQUVXZ"

func GetSymbolInfo(sym string) (SymbolInfo, error) {
	if res, ok := symInfos[sym]; ok {
		return *res, nil
	}
	return SymbolInfo{}, errors.New("no such symbol")
}

func newSymbolInfo(sym string) {
	sLen := len(sym)
	if sLen == 0 {
		return
	}
	if _, ok := symInfos[sym]; ok {
		return
	}
	var symInfo = SymbolInfo{}
	for i := 0; i < len(initTemp); i++ {
		if initTemp[i].TickerLen != sLen {
			continue
		}
		if sym[:len(initTemp[i].TickerPrefix)] != initTemp[i].TickerPrefix {
			continue
		}
		switch initTemp[i].DateLen {
		case 2:
			if sLen < 3 || !initTemp[i].USticker {
				continue
			}
			if sym[sLen-1] < '0' || sym[sLen-1] > '9' {
				continue
			}
			if mon := strings.IndexByte(usDeliverMonth, sym[sLen-2]); mon < 0 {
				// deliver Month no exist
				continue
			} else {
				symInfo.deliverMonth = mon
			}
		case 3:
			if sLen < 4 {
				continue
			}
			if res, err := strconv.Atoi(sym[sLen-2:]); err != nil {
				continue
			} else if res < 1 || res > 12 {
				continue
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
				continue
			} else {
				symInfo.deliverMonth = res
			}
		default:
			continue
		case 0:
		}
		symInfo.Ticker = sym
		symInfo.symbolBase = &initTemp[i].Base
		symInfos[sym] = &symInfo
		return
	}
}

func initSymbols() {
	if bb, err := ioutil.ReadFile("symbols.yml"); err == nil {
		var symTemps map[string]symbolTemplate
		if yaml.Unmarshal(bb, &symTemps) == nil {
			initTemp = []symbolTemplate{}
			for _, ss := range symTemps {
				initTemp = append(initTemp, ss)
			}
		}
	}
	sort.Sort(initTemp)
	// verify initTemp
	for i := 0; i < len(initTemp); i++ {
		if initTemp[i].Base.Margin <= 0 || initTemp[i].Base.Margin == 1.0 {
			initTemp[i].Base.bMargin = false
		} else {
			initTemp[i].Base.bMargin = true
		}
		if initTemp[i].Base.VolStep == 0 {
			initTemp[i].Base.VolStep = 1
		}
		if initTemp[i].Base.PriceStep == 0 {
			initTemp[i].Base.PriceStep = 1
		}
		if initTemp[i].Base.VolDigits > len(fMulti)-3 {
			initTemp[i].Base.VolDigits = len(fMulti) - 3
		}
		if initTemp[i].Base.PriceDigits > len(fMulti)-3 {
			initTemp[i].Base.PriceDigits = len(fMulti) - 3
		}
	}
}
