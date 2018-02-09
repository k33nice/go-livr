package livr

import (
	"errors"
	"reflect"
)

// anyObject - rule for checking that validated value is not empty object.
func anyObject(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		value := firstArg(builders...)
		if value == nil || value == "" {
			return value, nil
		}
		if reflect.ValueOf(value).Kind() != reflect.Map {
			return nil, errors.New("FORMAT_ERROR")
		}

		if isZero(reflect.ValueOf(value)) {
			return nil, errors.New("FORMAT_ERROR")
		}

		return value, nil
	}
}

// notEmpty - check that validated value is not empty if exists.
func notEmpty(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		value := firstArg(builders...)

		if value == nil {
			// TODO: return error
			return nil, nil
		}

		if isZero(reflect.ValueOf(value)) {
			return nil, errors.New("CANNOT_BE_EMPTY")
		}

		return value, nil
	}
}

// notEmptyList - check that validated value is not empty list.
func notEmptyList(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		value := firstArg(builders...)

		if value == nil || value == "" {
			return nil, errors.New("CANNOT_BE_EMPTY")
		}

		if reflect.TypeOf(value).Kind() != reflect.Array && reflect.TypeOf(value).Kind() != reflect.Slice {
			return nil, errors.New("FORMAT_ERROR")
		}

		if reflect.ValueOf(value).Len() == 0 {
			return nil, errors.New("CANNOT_BE_EMPTY")
		}

		return value, nil
	}
}

// required - checks that validated value exists and not empty.
func required(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		value := firstArg(builders...)
		if value == nil || value == "" {
			return nil, errors.New("REQUIRED")
		}
		return value, nil
	}
}
