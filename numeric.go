package livr

import (
	"errors"
	"strconv"
)

// decimal - check that validated value is decimal number.
func decimal(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case float64:
			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, nil
			}
			return nil, errors.New("NOT_DECIMAL")

		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// integer - make sure that validated value is integer.
func integer(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case float64:
			if v != float64(int(v)) {
				return nil, errors.New("NOT_INTEGER")
			}
			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				if f != float64(int(f)) {
					return nil, errors.New("NOT_INTEGER")
				}
				return f, nil
			}
			return nil, errors.New("NOT_INTEGER")
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// maxNumber - make sure that the validated value is not bigger than some number.
func maxNumber(args ...interface{}) Validation {
	var maxNumber float64
	if len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			maxNumber = v
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case float64:
			if v > maxNumber {
				return nil, errors.New("TOO_HIGH")
			}

			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				if f > maxNumber {
					return nil, errors.New("TOO_HIGH")
				}
				return f, nil
			}

			return nil, errors.New("NOT_NUMBER")
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// minNumber - make sure that the validated value is not lower than some number.
func minNumber(args ...interface{}) Validation {
	var minNumber float64
	if len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			minNumber = v
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case float64:
			if v < minNumber {
				return nil, errors.New("TOO_LOW")
			}

			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				if f < minNumber {
					return nil, errors.New("TOO_LOW")
				}
				return f, nil
			}

			return nil, errors.New("NOT_NUMBER")
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// numberBetween - make sure that validated value is number between minNumber and maxNumber.
func numberBetween(args ...interface{}) Validation {
	var minNumber, maxNumber float64
	if len(args) > 1 {
		if v, ok := args[0].(float64); ok {
			minNumber = v
		}
		if v, ok := args[1].(float64); ok {
			maxNumber = v
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case float64:
			if v > maxNumber {
				return nil, errors.New("TOO_HIGH")
			}
			if v < minNumber {
				return nil, errors.New("TOO_LOW")
			}

			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				if f > maxNumber {
					return nil, errors.New("TOO_HIGH")
				}
				if f < minNumber {
					return nil, errors.New("TOO_LOW")
				}
				return f, nil
			}

			return nil, errors.New("NOT_NUMBER")
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// positiveInteger - make sure that validated value is positive integer number.
func positiveInteger(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case float64:
			if v != float64(int(v)) || v <= 0 {
				return nil, errors.New("NOT_POSITIVE_INTEGER")
			}
			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				if f > 0 {
					return f, nil
				}
			}
			return nil, errors.New("NOT_POSITIVE_INTEGER")
		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// positiveDecimal - make sure that validated value is positive decimal number.
func positiveDecimal(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		switch v := value.(type) {
		case float64:
			if v <= 0 {
				return nil, errors.New("NOT_POSITIVE_DECIMAL")
			}
			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				if f > 0 {
					return f, nil
				}
			}
			return nil, errors.New("NOT_POSITIVE_DECIMAL")

		default:
			return nil, errors.New("FORMAT_ERROR")
		}
	}
}

// maxDecimals - make sure that validated value is a number with at most the specified number of
// digits after the decimal point, trailing zeros do not count.
func maxDecimals(args ...interface{}) Validation {
	args, _ = ruleArgs(args)
	var max int
	ok := len(args) > 0
	if ok {
		max, ok = countArg(args[0])
	}
	if !ok {
		schemaErrorf("Rule max_decimals takes a non-negative integer")
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		var decimals int
		var f float64
		switch v := value.(type) {
		case string:
			// Counted as written, so "0.10000000000000000001" has 20 digits.
			decimals = countDecimals(v)
			var err error
			if f, err = strconv.ParseFloat(v, 64); decimals < 0 || err != nil {
				return nil, errors.New("NOT_NUMBER")
			}
		case bool:
			return nil, errors.New("NOT_NUMBER")
		default:
			var ok bool
			if f, ok = number(v); !ok {
				return nil, errors.New("FORMAT_ERROR")
			}
			decimals = countDecimals(formatNumber(f))
		}

		if decimals > max {
			return nil, errors.New("TOO_MANY_DECIMALS")
		}
		return f, nil
	}
}
