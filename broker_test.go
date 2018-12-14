package ats

import (
	"reflect"
	"testing"
)

func TestRegisterBroker(t *testing.T) {
	type args struct {
		name string
		inf  Broker
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"RegisterBrokersim", args{"simBroker", &simTrader}, true},
		{"RegisterBrokersimDup", args{"simBroker1", &simTrader}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterBroker(tt.args.name, tt.args.inf); (err != nil) != tt.wantErr {
				t.Errorf("RegisterBroker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_OpenBroker(t *testing.T) {
	type args struct {
		name string
		ch   chan<- QuoteEvent
	}
	ch := make(chan QuoteEvent)
	simTrader.evChan = ch
	tests := []struct {
		name    string
		args    args
		want    Broker
		wantErr bool
	}{
		// TODO: Add test cases.
		{"openBroker1", args{"simBroker", ch}, &simTrader, false},
		{"openBroker2", args{"simBroker2", ch}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openBroker(tt.args.name, tt.args.ch)
			if (err != nil) != tt.wantErr {
				t.Errorf("openBroker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("openBroker() = %v, want %v", got, tt.want)
			}
		})
	}
}
