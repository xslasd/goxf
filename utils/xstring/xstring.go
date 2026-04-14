package xstring

import "github.com/goccy/go-json"

func Json(obj any) string {
	str, _ := json.Marshal(obj)
	return string(str)
}
