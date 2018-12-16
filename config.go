package ats

type Config map[string]interface{}

func (c Config) Put(key string, v interface{}) {
	c[key] = v
}

func (c Config) GetInt(key string, def int) int {
	if r, ok := c[key]; ok {
		if res, ok := r.(int); ok {
			return res
		}
	}
	return def
}

func (c Config) GetString(key string, def string) string {
	if r, ok := c[key]; ok {
		if res, ok := r.(string); ok {
			return res
		}
	}
	return def
}

func (c Config) GetStrings(key string) []string {
	if r, ok := c[key]; ok {
		if res, ok := r.([]string); ok {
			return res
		}
	}
	return []string{}
}

func (c Config) GetFloat64(key string, def float64) float64 {
	if r, ok := c[key]; ok {
		if res, ok := r.(float64); ok {
			return res
		}
	}
	return def
}
