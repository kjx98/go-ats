package ats

import (
	"reflect"
	"testing"
)

var conf = Config{}

func TestConfig_Put(t *testing.T) {
	type args struct {
		key string
		v   interface{}
	}

	tests := []struct {
		name string
		c    Config
		args args
	}{
		// TODO: Add test cases.
		{"Put1", conf, args{"name", "test1"}},
		{"Put2", conf, args{"value1", 123}},
		{"Put3", conf, args{"value2", 567.98}},
		{"Put4", conf, args{"strings", []string{"test1", "test2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Put(tt.args.key, tt.args.v)
		})
	}
}

func TestConfig_GetInt(t *testing.T) {
	type args struct {
		key string
		def int
	}
	tests := []struct {
		name string
		c    Config
		args args
		want int
	}{
		// TODO: Add test cases.
		{"Geti1", conf, args{"name", 100}, 100},
		{"Geti2", conf, args{"value1", 100}, 123},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.GetInt(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("Config.GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetString(t *testing.T) {
	type args struct {
		key string
		def string
	}
	tests := []struct {
		name string
		c    Config
		args args
		want string
	}{
		// TODO: Add test cases.
		{"gets1", conf, args{"name", "test"}, "test1"},
		{"gets2", conf, args{"value1", "100"}, "100"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.GetString(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("Config.GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetFloat64(t *testing.T) {
	type args struct {
		key string
		def float64
	}
	tests := []struct {
		name string
		c    Config
		args args
		want float64
	}{
		// TODO: Add test cases.
		{"Getf1", conf, args{"name", 100}, 100},
		{"Getf2", conf, args{"value1", 100}, 100},
		{"Getf3", conf, args{"value2", 200}, 567.98},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.GetFloat64(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("Config.GetFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetStrings(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		c    Config
		args args
		want []string
	}{
		// TODO: Add test cases.
		{"getss1", conf, args{"name"}, []string{}},
		{"getss2", conf, args{"value1"}, []string{}},
		{"getss3", conf, args{"strings"}, []string{"test1", "test2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.GetStrings(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.GetStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
