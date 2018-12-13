package ats

import (
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterBroker(tt.args.name, tt.args.inf); (err != nil) != tt.wantErr {
				t.Errorf("RegisterBroker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
