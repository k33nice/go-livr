package livr

import (
	"errors"
	"reflect"
)

// AnyObject - rule for checking that validated value is not empty object.
func AnyObject(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
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

// NotEmpty - check that validated value is not empty if exists.
func NotEmpty(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}

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

// NotEmptyList - check that validated value is not empty list.
func NotEmptyList(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}

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

// Required - checks that validated value exists and not empty.
func Required(...interface{}) func(...interface{}) (interface{}, interface{}) {
	return func(builders ...interface{}) (interface{}, interface{}) {
		var value interface{}
		if len(builders) > 0 {
			value = builders[0]
		}
		if value == nil || value == "" {
			return nil, errors.New("REQUIRED")
		}
		return value, nil
	}
}
