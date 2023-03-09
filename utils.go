package main

import "math"

func ToInt(v interface{}) (r int) {
	switch t := v.(type) {
	default:
		r = 0
	case float64:
		r = int(math.Round(t))
	case float32:
		r = ToInt(float64(t))
	case int64:
		r = int(t)
	case int32:
		r = int(t)
	case uint:
		r = int(t)
	case uint64:
		r = int(t)
	case uint32:
		r = int(t)
	case int:
		r = v.(int)
	}
	return
}
