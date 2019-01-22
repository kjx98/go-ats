package ats

import (
	"reflect"
	"testing"
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
		{"LoadRunTick-EUR", args{"EURUSD"}, 53803847, false},
		{"LoadRunTick-GBP", args{"GBPUSD"}, 2129280, false},
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
		evCh := make(chan QuoteEvent)
		defer close(evCh)
		bb, err := simTrader.Open(evCh)
		if err != nil {
			t.Error("simBroker Open", err)
			return
		}
		b = bb.(simBroker)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.SendOrder(tt.args.sym, tt.args.dir, tt.args.qty, tt.args.prc, tt.args.stopL); got != tt.want {
				t.Errorf("simBroker.SendOrder() = %v, want %v", got, tt.want)
			}
		})
	}
	dumpOrderBook("EURUSD")
}

func Test_simBroker_CancelOrder(t *testing.T) {
	type args struct {
		oid int
	}
	if b < 0 {
		evCh := make(chan QuoteEvent)
		defer close(evCh)
		bb, err := simTrader.Open(evCh)
		if err != nil {
			t.Error("simBroker Open", err)
			return
		}
		b = bb.(simBroker)
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
	dumpOrderBook("EURUSD")
}
