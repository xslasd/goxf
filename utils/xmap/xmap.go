package xmap

import (
	"fmt"
	"reflect"
)

// MergeStringMap 合并两个字符串键的映射 src->dest
func MergeStringMap(dest, src map[string]any) {
	for sk, sv := range src {
		tv, ok := dest[sk]
		if !ok {
			dest[sk] = sv
			continue
		}
		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)
		if svType != tvType {
			continue
		}

		switch ttv := tv.(type) {
		case map[any]any:
			tsv := sv.(map[any]any)
			ssv := ToMapStringInterface(tsv)
			stv := ToMapStringInterface(ttv)
			MergeStringMap(stv, ssv)
			dest[sk] = stv
		case map[string]any:
			MergeStringMap(ttv, sv.(map[string]any))
			dest[sk] = ttv
		default:
			dest[sk] = sv
		}
	}
}

// MergeStringMapWithChanged 合并两个字符串键的映射 src->dest，并记录变化
func MergeStringMapWithChanged(dest, src, change map[string]any, prefix string) {
	for sk, sv := range src {
		tv, ok := dest[sk]
		if !ok {
			dest[sk] = sv
			continue
		}
		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)
		if svType != tvType {
			continue
		}
		pp := fmt.Sprintf("%s%s%s", prefix, ".", sk)
		if prefix == "" {
			pp = sk
		}

		switch ttv := tv.(type) {
		case map[any]any:
			tsv := sv.(map[any]any)
			ssv := ToMapStringInterface(tsv)
			stv := ToMapStringInterface(ttv)
			MergeStringMapWithChanged(stv, ssv, change, pp)
			dest[sk] = stv
		case map[string]any:
			MergeStringMapWithChanged(ttv, sv.(map[string]any), change, pp)
			dest[sk] = ttv
		default:
			if !reflect.DeepEqual(dest[sk], sv) {
				change[pp] = sv
			}
			dest[sk] = sv
		}
	}
}

// ToMapStringInterface 将 map[any]any 转换为 map[string]any
func ToMapStringInterface(src map[any]any) map[string]any {
	tgt := map[string]any{}
	for k, v := range src {
		tgt[fmt.Sprintf("%v", k)] = v
	}
	return tgt
}
