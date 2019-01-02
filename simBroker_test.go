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
				t.Errorf("LoadRunTick() = %d, want %d", got, tt.want)
			}
		})
	}
}
