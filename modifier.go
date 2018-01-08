package livr

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Default -
func Default(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var defVal interface{}
	if len(args) > 0 {
		defVal = args[0]
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var val interface{}
		if len(builders) > 0 {
			val = builders[0]
		}

		if val == nil || val == "" {
			return defVal, nil
		}

		return val, nil
	}
}

// Trim -
func Trim(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var val interface{}
		if len(builders) > 0 {
			val = builders[0]
		}

		if val == nil || val == "" {
			return val, nil
		}

		switch v := val.(type) {
		case []string:
			var results []string
			for _, s := range v {
				s = strings.Trim(s, " ")
				results = append(results, s)
			}
			return results, nil
		case map[string]string:
			var results = make(map[string]string)
			for i, s := range v {
				s = strings.Trim(s, " ")
				results[i] = s
			}
			return results, nil
		case string:
			return strings.Trim(v, " "), nil
		default:
			return v, nil
		}
	}
}

// ToLc -
func ToLc(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var val interface{}
		if len(builders) > 0 {
			val = builders[0]
		}

		if val == nil || val == "" {
			return val, nil
		}

		v := reflect.ValueOf(val)
		switch v.Kind() {
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetString(strings.ToLower(v.Index(i).String()))
			}
			return v.Interface(), nil
		case reflect.Map:
			for _, k := range v.MapKeys() {
				lcVal := strings.ToLower(v.MapIndex(k).Interface().(string))
				v.SetMapIndex(k, reflect.ValueOf(lcVal))
			}
			return v.Interface(), nil
		case reflect.String:
			return strings.ToLower(v.String()), nil
		default:
			return val, nil
		}
	}
}

// ToUc -
func ToUc(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var val interface{}
		if len(builders) > 0 {
			val = builders[0]
		}

		if val == nil || val == "" {
			return val, nil
		}

		v := reflect.ValueOf(val)
		switch v.Kind() {
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetString(strings.ToUpper(v.Index(i).String()))
			}
			return v.Interface(), nil
		case reflect.Map:
			for _, k := range v.MapKeys() {
				lcVal := strings.ToUpper(v.MapIndex(k).Interface().(string))
				v.SetMapIndex(k, reflect.ValueOf(lcVal))
			}
			return v.Interface(), nil
		case reflect.String:
			return strings.ToUpper(v.String()), nil
		default:
			return val, nil
		}
	}
}

// Remove -
func Remove(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var chars string
	if len(args) > 0 {
		if v, ok := args[0].(string); ok {
			chars = v
		}
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var val interface{}
		if len(builders) > 0 {
			val = builders[0]
		}

		re, err := regexp.Compile(fmt.Sprintf("[%s]", strings.Replace(regexp.QuoteMeta(chars), "-", `\-`, -1)))
		if err != nil {
			return val, nil
		}

		switch s := val.(type) {
		case string:
			return re.ReplaceAllString(s, ""), nil
		case float64:
			newS := re.ReplaceAllString(strconv.FormatFloat(s, 'f', -1, 64), "")
			if r, err := strconv.ParseFloat(newS, 64); err == nil {
				return r, nil
			}
			return newS, nil
		case bool:
			newS := re.ReplaceAllString(strconv.FormatBool(s), "")
			return newS, nil
		default:
			return val, nil
		}
	}
}

// LeaveOnly -
func LeaveOnly(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var chars string
	if len(args) > 0 {
		if v, ok := args[0].(string); ok {
			chars = v
		}
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var val interface{}
		if len(builders) > 0 {
			val = builders[0]
		}

		re, err := regexp.Compile(fmt.Sprintf("[^%s]", strings.Replace(regexp.QuoteMeta(chars), "-", `\-`, -1)))
		if err != nil {
			return val, nil
		}

		switch s := val.(type) {
		case string:
			return re.ReplaceAllString(s, ""), nil
		case float64:
			newS := re.ReplaceAllString(strconv.FormatFloat(s, 'f', -1, 64), "")
			if r, err := strconv.ParseFloat(newS, 64); err == nil {
				return r, nil
			}
			return newS, nil
		case bool:
			newS := re.ReplaceAllString(strconv.FormatBool(s), "")
			return newS, nil
		default:
			return val, nil
		}
	}
}
