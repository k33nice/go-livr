package livr

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// _default - set default for validated value.
func _default(args ...interface{}) Validation {
	defVal := firstArg(args...)

	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
		if val == nil || val == "" {
			return defVal, nil
		}

		return val, nil
	}
}

// trim - trim spaces from validated value.
func trim(args ...interface{}) Validation {
	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
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

// toLc - convert validated value to lower case.
func toLc(args ...interface{}) Validation {
	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
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

// toUc - convert validated value to upper case.
func toUc(args ...interface{}) Validation {
	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
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

// remove - remove specified characters from value.
func remove(args ...interface{}) Validation {
	var chars string
	if len(args) > 0 {
		if v, ok := args[0].(string); ok {
			chars = v
		}
	}

	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
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

// leaveOnly - leave only specified characters in value.
func leaveOnly(args ...interface{}) Validation {
	var chars string
	if len(args) > 0 {
		if v, ok := args[0].(string); ok {
			chars = v
		}
	}

	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
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
