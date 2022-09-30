package filter

import (
	"fmt"
	"strings"
)

type Filter []string

func Parse(f string) Filter {
	return Filter(strings.Split(f, " "))
}

func ParseMany(t interface{}) []Filter {
	switch e := t.(type) {
	case map[string]interface{}:
		ret := []Filter{}
		for base, v := range e {
			f := ParseMany(v)
			for _, ff := range f {
				ret = append(ret, append([]string{base}, ff...))
			}
		}
		return ret
	case []interface{}:
		ret := []Filter{}
		for _, v := range e {
			f := ParseMany(v)
			ret = append(ret, f...)
		}
		return ret
	case string:
		return []Filter{Parse(e)}
	default:
		panic(fmt.Sprintf("filter.ParseMany: unknown type \"%T\"", t))
	}
}
