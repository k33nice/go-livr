package livr

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// eq - make sure that validated value is equal to some value.
func eq(args ...interface{}) Validation {
	allowed := firstArg(args...)

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return nil, nil
		}

		switch value.(type) {
		case float64, string, bool:
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
		if fmt.Sprint(value) != fmt.Sprint(allowed) {
			return nil, errors.New("NOT_ALLOWED_VALUE")
		}

		return value, nil
	}
}

// lengthBetween - make sure that validated value's length is between specified length.
func lengthBetween(args ...interface{}) Validation {
	var minLength, maxLength float64
	if len(args) > 1 {
		if v, ok := args[0].(float64); ok {
			minLength = v
		}
		if v, ok := args[1].(float64); ok {
			maxLength = v
		}
	}
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case string:
			if utf8.RuneCountInString(v) > int(maxLength) {
				return nil, errors.New("TOO_LONG")
			}
			if utf8.RuneCountInString(v) < int(minLength) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		case bool:
			if utf8.RuneCountInString(strconv.FormatBool(v)) > int(maxLength) {
				return nil, errors.New("TOO_LONG")
			}
			if utf8.RuneCountInString(strconv.FormatBool(v)) < int(minLength) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		case float64:
			if utf8.RuneCountInString(strconv.FormatFloat(v, 'f', -1, 64)) > int(maxLength) {
				return nil, errors.New("TOO_LONG")
			}
			if utf8.RuneCountInString(strconv.FormatFloat(v, 'f', -1, 64)) < int(minLength) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// lengthEqual - make sure that validated value's length is equal to specified length.
func lengthEqual(args ...interface{}) Validation {
	var length float64
	if len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			length = v
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case string:
			if utf8.RuneCountInString(v) > int(length) {
				return nil, errors.New("TOO_LONG")
			}
			if utf8.RuneCountInString(v) < int(length) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		case bool:
			if utf8.RuneCountInString(strconv.FormatBool(v)) > int(length) {
				return nil, errors.New("TOO_LONG")
			}
			if utf8.RuneCountInString(strconv.FormatBool(v)) < int(length) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		case float64:
			if utf8.RuneCountInString(strconv.FormatFloat(v, 'f', -1, 64)) > int(length) {
				return nil, errors.New("TOO_LONG")
			}
			if utf8.RuneCountInString(strconv.FormatFloat(v, 'f', -1, 64)) < int(length) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// like - check that validated value is like specified regexp.
func like(args ...interface{}) Validation {
	var re *regexp.Regexp
	var flags string
	if len(args) > 0 {
		if len(args) > 1 {
			if v, ok := args[1].(string); ok {
				if v == "i" {
					flags = "(?i)"
				}
			}
		}

		if v, ok := args[0].(string); ok {
			reg, err := regexp.Compile(flags + v)
			if err != nil {
				re = regexp.MustCompile(".*")
			} else {
				re = reg
			}
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case string:
			if ok := re.MatchString(v); !ok {
				return nil, errors.New("WRONG_FORMAT")
			}
			return v, nil
		case float64:
			if ok := re.MatchString(strconv.FormatFloat(v, 'f', -1, 64)); !ok {
				return nil, errors.New("WRONG_FORMAT")
			}
			return v, nil
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// maxLength - check that validated value's length is not longer than specified.
func maxLength(args ...interface{}) Validation {
	var minLength float64
	if len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			minLength = v
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case string:
			if utf8.RuneCountInString(v) > int(minLength) {
				return nil, errors.New("TOO_LONG")
			}
			return v, nil
		case bool:
			if utf8.RuneCountInString(strconv.FormatBool(v)) > int(minLength) {
				return nil, errors.New("TOO_LONG")
			}
			return v, nil
		case float64:
			if utf8.RuneCountInString(strconv.FormatFloat(v, 'f', -1, 64)) > int(minLength) {
				return nil, errors.New("TOO_LONG")
			}
			return v, nil
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// minLength - check that validated value's length is not shorter than specified.
func minLength(args ...interface{}) Validation {
	var minLength float64
	if len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			minLength = v
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case string:
			if utf8.RuneCountInString(v) < int(minLength) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		case bool:
			if utf8.RuneCountInString(strconv.FormatBool(v)) < int(minLength) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		case float64:
			if utf8.RuneCountInString(strconv.FormatFloat(v, 'f', -1, 64)) < int(minLength) {
				return nil, errors.New("TOO_SHORT")
			}
			return v, nil
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// oneOf - check that validated value is one of specified.
func oneOf(args ...interface{}) Validation {
	var allowed []interface{}
	if len(args) > 0 {
		if v, ok := args[0].([]interface{}); ok {
			allowed = v
		} else {
			allowed = args[0 : len(args)-1]
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		for _, val := range allowed {
			switch v := val.(type) {
			case bool:
				// A boolean matches a string in the forms of strconv.ParseBool, "true", "1",
				// "t", "TRUE" and so on, as before LIVR 2.5.
				switch vv := value.(type) {
				case bool:
					if vv == v {
						return val, nil
					}
				case string:
					if vs, err := strconv.ParseBool(vv); err == nil && vs == v {
						return val, nil
					}
				default:
					if _, ok := stringForm(value); !ok {
						return nil, errors.New("FORMAT_ERROR")
					}
				}
			case string:
				switch vv := value.(type) {
				case bool:
					if vb := strconv.FormatBool(vv); vb == v {
						return val, nil
					}
				case string:
					if vv == v {
						return val, nil
					}
				case float64:
					if vf := strconv.FormatFloat(vv, 'f', -1, 64); vf == v {
						return val, nil
					}
				default:
					return nil, errors.New("FORMAT_ERROR")
				}
			case float64:
				switch vv := value.(type) {
				case string:
					if vs := strconv.FormatFloat(v, 'f', -1, 64); vs == vv {
						return val, nil
					}
				case float64:
					if vv == v {
						return val, nil
					}
				default:
					return nil, errors.New("FORMAT_ERROR")
				}
			}
		}
		return nil, errors.New("NOT_ALLOWED_VALUE")
	}
}

// _string - check that validated value is string-like.
func _string(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case string:
			return v, nil
		case bool:
			return strconv.FormatBool(v), nil
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// notOneOf - check that validated value is none of specified.
func notOneOf(args ...interface{}) Validation {
	args, _ = ruleArgs(args)
	if len(args) == 0 {
		schemaErrorf("Rule not_one_of takes one or more values")
	}

	forbidden := make([]string, 0, len(args))
	for _, arg := range args {
		s, ok := stringForm(arg)
		if !ok {
			schemaErrorf("Rule not_one_of takes primitive values")
		}
		forbidden = append(forbidden, s)
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		s, ok := stringForm(value)
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		if containsString(forbidden, s) {
			return nil, errors.New("NOT_ALLOWED_VALUE")
		}

		return value, nil
	}
}

// digits - check that validated value is one or more digits 0-9, and convert it to a string.
func digits(args ...interface{}) Validation {
	noArgs("digits", args)

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		s, ok := stringForm(value)
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		for i := 0; i < len(s); i++ {
			if !isDigit(s[i]) {
				return nil, errors.New("NOT_DIGITS")
			}
		}

		return s, nil
	}
}

// startsWith - check that validated value starts with one of specified strings.
func startsWith(args ...interface{}) Validation {
	return affixRule("starts_with", args, strings.HasPrefix)
}

// endsWith - check that validated value ends with one of specified strings.
func endsWith(args ...interface{}) Validation {
	return affixRule("ends_with", args, strings.HasSuffix)
}

// contains - check that validated value contains one of specified strings.
func contains(args ...interface{}) Validation {
	return affixRule("contains", args, strings.Contains)
}

// affixRule - check that the string form of validated value matches one of the arguments.
func affixRule(name string, args []interface{}, matches func(s, affix string) bool) Validation {
	args, _ = ruleArgs(args)
	if len(args) == 0 {
		schemaErrorf("Rule %s takes one or more strings", name)
	}

	affixes := make([]string, 0, len(args))
	for _, arg := range args {
		s, ok := stringArg(arg)
		if !ok || s == "" {
			schemaErrorf("Rule %s takes non-empty strings", name)
		}
		affixes = append(affixes, s)
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		s, ok := stringForm(value)
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		for _, affix := range affixes {
			if matches(s, affix) {
				return s, nil
			}
		}

		return nil, errors.New("WRONG_FORMAT")
	}
}
