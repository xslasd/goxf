package xmap

import (
	"fmt"
	"testing"
)

func TestMergeStringMapWithChanged(t *testing.T) {
	src := map[string]any{
		"name": any("richard"),
		"age":  any(18),
		"nest": map[string]any{
			"color":  any("red"),
			"weight": any(65.5),
		},
	}

	dest := map[string]any{
		"name": any("richard"),
		"age":  any(38),
		"nest": map[string]any{
			"color":  any("blue"),
			"weight": any(65.5),
		},
	}

	changed := make(map[string]any)

	MergeStringMapWithChanged(dest, src, changed, "")

	fmt.Printf("changed: %+v", changed)

	fmt.Printf("src: %+v", src)

	fmt.Printf("dest: %+v", dest)

	if len(changed) != 2 {
		t.Errorf("not enougth changed value!")
	}
}
