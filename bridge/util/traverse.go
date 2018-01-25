package util

import (
	"strings"

	"github.com/orktes/homeautomation/bridge/adapter"
)

func Traverse(vc adapter.ValueContainer, cb func(key string, val interface{}) error, includeRoot bool) error {
	var iterate func(vc adapter.ValueContainer, prefix []string) error
	iterate = func(vc adapter.ValueContainer, prefix []string) error {
		vals, err := vc.GetAll()
		if err != nil {
			return err
		}

		for key, val := range vals {
			fullkeysparts := make([]string, len(prefix)+1)
			copy(fullkeysparts, prefix)
			fullkeysparts[len(fullkeysparts)-1] = key

			switch val := val.(type) {
			case adapter.ValueContainer:
				if err := iterate(val, fullkeysparts); err != nil {
					return err
				}
			default:
				if err := cb(strings.Join(fullkeysparts, "/"), val); err != nil {
					return err
				}
			}
		}

		return nil
	}

	prefix := []string{}

	if includeRoot {
		if ad, ok := vc.(adapter.Adapter); ok {
			prefix = append(prefix, ad.ID())
		}
	}

	return iterate(vc, prefix)
}
