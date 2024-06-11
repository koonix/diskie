package main

import (
	"html/template"
	"reflect"
	"regexp"

	"github.com/dustin/go-humanize"
)

var condenseSpaceRegex, condenseDashRegex *regexp.Regexp

func init() {
	condenseSpaceRegex = regexp.MustCompile(`\s+`)
	condenseDashRegex = regexp.MustCompile(`-+`)
}

var templateFuncs = template.FuncMap{

	"dereference": func(v any) any {
		vv := reflect.ValueOf(v)
		if vv.Kind() == reflect.Ptr {
			if vv.IsNil() {
				return reflect.Zero(vv.Type().Elem()).Interface()
			} else {
				return vv.Elem().Interface()
			}
		} else {
			return v
		}
	},

	"isEmpty": func(v any) bool {
		if v == nil {
			return true
		}
		switch vv := v.(type) {
		case string:
			return vv == ""
		case *string:
			return *vv == ""
		default:
			return false
		}
	},

	"condense": func(v string) string {
		v = condenseSpaceRegex.ReplaceAllString(v, " ")
		v = condenseDashRegex.ReplaceAllString(v, "-")
		return v
	},

	"humanBytes": func(v uint64) string {
		return humanize.Bytes(v)
	},

	"humanBytesIEC": func(v uint64) string {
		return humanize.IBytes(v)
	},
}
