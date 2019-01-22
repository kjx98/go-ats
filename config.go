package ats

import "reflect"

// Config map string to empty interface
type Config map[string]interface{}

// Put store config value
func (c Config) Put(key string, v interface{}) {
	c[key] = v
}

// GetInt ... get int config value
func (c Config) GetInt(key string, def int) int {
	if r, ok := c[key]; ok {
		switch r.(type) {
		case int8, int16, int32, int, int64:
			return int(reflect.ValueOf(r).Int())
		}
	}
	return def
}

// GetString ...	get string config value
func (c Config) GetString(key string, def string) string {
	if r, ok := c[key]; ok {
		if res, ok := r.(string); ok {
			return res
		}
	}
	return def
}

// GetStrings ...	get []string config value
func (c Config) GetStrings(key string) []string {
	if r, ok := c[key]; ok {
		if res, ok := r.([]string); ok {
			return res
		}
	}
	return []string{}
}

// GetFloat64 ...	get flota64 config value
func (c Config) GetFloat64(key string, def float64) float64 {
	if r, ok := c[key]; ok {
		if res, ok := r.(float64); ok {
			return res
		}
	}
	return def
}
