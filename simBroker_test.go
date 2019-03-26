package ats

import (
	"math/rand"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/kjx98/golib/julian"
)

func TestLoadRunTick(t *testing.T) {
	type args struct {
		sym string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		//{"LoadRunTick-EUR", args{"EURUSD"}, 2127996, false},
		//{"LoadRunTick-EUR", args{"EURUSD"}, 53803847, false},
		{"LoadRunTick-EUR", args{"EURUSD"}, 53803921, false},
		//{"LoadRunTick-GBP", args{"GBPUSD"}, 2129280, false},
		//{"LoadRunTick-GBP", args{"GBPUSD"}, 47765457, false},
		{"LoadRunTick-GBP", args{"GBPUSD"}, 47765531, false},
		{"LoadRunTick-JPY", args{"USDJPY"}, 2126252, false},
		{"LoadRunTick-XAU", args{"XAUUSD"}, 2009644, false},
		{"LoadRunTick-ETF", args{"sh510500"}, 5648, false},
		{"LoadRunTick-601318", args{"sh601318"}, 11268, false},
	}
	if noDukasData {
		t.Log("no tickData")
		return
	}
	if len(symbolsMap) == 0 {
		tests[4].want = 0
		tests[4].wantErr = true
		tests[5].want = 0
		tests[5].wantErr = true
	}
	simLoadSymbols()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadRunTick(tt.args.sym)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRunTick() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got.Len(), tt.want) {
				t.Errorf("LoadRunTick() = %d, want %d", got.Len(), tt.want)
			}
		})
	}
}

var subs = []QuoteSubT{}

func TestValidateTick(t *testing.T) {
	type args struct {
		sym string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"LoadRunTick-EUR", args{"EURUSD"}, false},
		{"LoadRunTick-GBP", args{"GBPUSD"}, false},
		{"LoadRunTick-JPY", args{"USDJPY"}, false},
		{"LoadRunTick-XAU", args{"XAUUSD"}, false},
		{"LoadRunTick-ETF", args{"sh510500"}, false},
		{"LoadRunTick-601318", args{"sh601318"}, false},
	}
	if noDukasData {
		t.Log("no tickData")
		return
	}
	if len(symbolsMap) == 0 {
		tests[4].wantErr = true
		tests[5].wantErr = true
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTick(tt.args.sym); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTick() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if si, err := GetSymbolInfo(tt.args.sym); err == nil {
					var subo = QuoteSubT{Symbol: tt.args.sym}
					subo.QuotesPtr = si.getQuotesPtr()
					subs = append(subs, subo)
				}
			}
		})
	}
}

var b simBroker = -1

func Test_simBroker_SendOrder(t *testing.T) {
	type args struct {
		sym   string
		dir   OrderDirT
		qty   int
		prc   float64
		stopL float64
	}
	if b < 0 {
		if bb, err := simTrader.Open(nil); err == nil {
			b = bb.(simBroker)
		} else {
			t.Error("simBroker Open", err)
			return
		}
	}

	tests := []struct {
		name string
		b    simBroker
		args args
		want int
	}{
		// TODO: Add test cases.
		{"SendOrder1", b, args{"EURUSD", OrderDirBuy, 1, 1.1355, 0}, 1},
		{"SendOrder2", b, args{"EURUSD", OrderDirBuy, 1, 1.1358, 0}, 2},
		{"SendOrder3", b, args{"EURUSD", OrderDirBuy, 2, 1.1355, 0}, 3},
		{"SendOrder4", b, args{"EURUSD", OrderDirSell, 1, 1.1365, 0}, 4},
		{"SendOrder5", b, args{"EURUSD", OrderDirSell, 1, 1.1362, 0}, 5},
		{"SendOrder6", b, args{"EURUSD", OrderDirSell, 2, 1.1365, 0}, 6},
	}
	if noDukasData {
		t.Log("no tickData")
		return
	}
	bSimValidate = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.SendOrder(tt.args.sym, tt.args.dir, tt.args.qty, tt.args.prc, tt.args.stopL); got != tt.want {
				t.Errorf("simBroker.SendOrder() = %v, want %v", got, tt.want)
			}
		})
	}
	dumpSimOrderBook("EURUSD")
}

func Test_simBroker_CancelOrder(t *testing.T) {
	type args struct {
		oid int
	}
	if b < 0 {
		if bb, err := simTrader.Open(nil); err == nil {
			b = bb.(simBroker)
		} else {
			t.Error("simBroker Open", err)
			return
		}
	}
	tests := []struct {
		name    string
		b       simBroker
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"CancelOrder1", b, args{1}, false},
		{"CancelOrder2", b, args{0}, true},
		{"CancelOrder3", b, args{1}, true},
		{"CancelOrder4", b, args{5}, false},
		{"CancelOrder5", b, args{5}, true},
		{"CancelOrder6", b, args{4}, false},
	}
	if noDukasData {
		t.Log("no tickData")
		return
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.CancelOrder(tt.args.oid); (err != nil) != tt.wantErr {
				t.Errorf("simBroker.CancelOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	dumpSimOrderBook("EURUSD")
}

const nBrokers = 64

func Test_simBroker_Start(t *testing.T) {
	type args struct {
		c Config
	}
	if b < 0 {
		if bb, err := simTrader.Open(nil); err == nil {
			b = bb.(simBroker)
		} else {
			t.Error("simBroker Open", err)
			return
		}
	}
	bSimValidate = false
	var bs = [nBrokers]simBroker{}
	bs[0] = b
	for i := 1; i < nBrokers; i++ {
		if bb, err := simTrader.Open(nil); err == nil {
			b0 := bb.(simBroker)
			bs[i] = b0
		}
	}
	// prepare 1e6 orders
	// odd for bs[0], even for bs[1]
	// price 3100 .. 51000 * 0.00001 + 1.0
	// EUR 1.031 .. 1.163
	// GBP 1.19 .. 1.51

	var sym string
	var pr float64
	var dir OrderDirT
	for i := 0; i < 5e6; i++ {
		br := bs[i&(nBrokers-1)]
		pB := rand.Intn(48000)
		vol := rand.Intn(10) + 1
		pr = 1.03 + float64(pB)*0.00001
		if pB&1 != 0 {
			// GBP
			sym = "GBPUSD"
			if pr <= 1.35 {
				dir = OrderDirBuy
			} else {
				dir = OrderDirSell
			}
		} else {
			// EUR
			sym = "EURUSD"
			if pr <= 1.097 {
				dir = OrderDirBuy
			} else {
				dir = OrderDirSell
			}
		}
		br.SendOrder(sym, dir, vol, pr, 0)
	}

	tests := []struct {
		name    string
		b       simBroker
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"simBroker-Start", b, args{Config{}}, false},
	}
	st1 := julian.FromUint32(20050301)
	en1 := julian.FromUint32(20181231)
	startTime = timeT64FromTime(st1.UTC())
	endTime = timeT64FromTime(en1.UTC())
	if len(subs) > 0 {
		if err := b.SubscribeQuotes(subs); err != nil {
			log.Error("Broker SubscribeQuotes", err)
			return
		}
	}
	dumpSimOrderStats()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.Start(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("simBroker.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	for atomic.LoadInt32(&simStatus) == VmRunning {
		runtime.Gosched()
	}
	dumpSimOrderStats()
	dumpSimBroker()
}
