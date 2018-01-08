package livr

import (
	"errors"
	"strconv"
)

// Decimal - check that validated value is decimal number.
func Decimal(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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

// Integer - make sure that validated value is integer.
func Integer(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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

// MaxNumber - make sure that the validated value is not bigger than some number.
func MaxNumber(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var maxNumber float64
	if len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			maxNumber = v
		}
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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

// MinNumber - make sure that the validated value is not lower than some number.
func MinNumber(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var minNumber float64
	if len(args) > 0 {
		if v, ok := args[0].(float64); ok {
			minNumber = v
		}
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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

// NumberBetween - make sure that validated value is number between minNumber and maxNumber.
func NumberBetween(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var minNumber, maxNumber float64
	if len(args) > 1 {
		if v, ok := args[0].(float64); ok {
			minNumber = v
		}
		if v, ok := args[1].(float64); ok {
			maxNumber = v
		}
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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

// PositiveInteger - make sure that validated value is positive integer number.
func PositiveInteger(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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

// PositiveDecimal - make sure that validated value is positive decimal number.
func PositiveDecimal(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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
