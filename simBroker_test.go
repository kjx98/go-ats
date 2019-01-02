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
		{"LoadRunTick-EUR", args{"EURUSD"}, 2127996, false},
		{"LoadRunTick-GBP", args{"GBPUSD"}, 2129280, false},
		{"LoadRunTick-JPY", args{"USDJPY"}, 2126252, false},
		{"LoadRunTick-XAU", args{"XAUUSD"}, 2009644, false},
		{"LoadRunTick-ETF", args{"sh510500"}, 5648, false},
		{"LoadRunTick-601318", args{"sh601318"}, 11268, false},
	}
	if noDukasData {
		t.Log("no FX data")
		return
	}
	if len(symbolsMap) == 0 {
		t.Log("no mysql connection")
		return
	}
	simLoadSymbols()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRunTick(tt.args.sym)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRunTick() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Len(), tt.want) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTick(tt.args.sym); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTick() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
