package livr

import (
	"errors"
	"reflect"
)

// anyObject - rule for checking that validated value is not empty object.
func anyObject(...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
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
func notEmpty(...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
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
func notEmptyList(...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {

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
func required(...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return nil, errors.New("REQUIRED")
		}
		return value, nil
	}
}

// requiredIf - checks that validated value is not empty when other fields have some values:
// {"required_if": {"type": "company", "address/country": ["UA", "PL"]}}.
func requiredIf(args ...interface{}) Validation {
	args, _ = ruleArgs(args)
	var conditions Dictionary
	if len(args) > 0 {
		conditions, _ = args[0].(Dictionary)
	}
	if len(conditions) == 0 {
		schemaErrorf("Rule required_if takes an object of references to values")
	}

	// The expected values as strings, so that 1 matches "1" and true matches "true".
	var paths [][]string
	var expected [][]string
	for reference, value := range conditions {
		paths = append(paths, parseReference(reference, "required_if"))

		list, ok := value.([]interface{})
		if !ok {
			list = []interface{}{value}
		} else if len(list) == 0 {
			schemaErrorf("Rule required_if got an empty list of values for %q", reference)
		}

		var strValues []string
		for _, v := range list {
			s, ok := stringForm(v)
			if !ok || s == "" {
				schemaErrorf("Rule required_if takes primitive values other than \"\" for %q", reference)
			}
			strValues = append(strValues, s)
		}
		expected = append(expected, strValues)
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if !isEmpty(value) {
			return value, nil
		}

		// The field is required when every referenced value is one of the expected values.
		params := containingObject(builders)
		for i, path := range paths {
			refValue, ok := stringForm(resolveReference(params, path))
			if !ok || refValue == "" || !containsString(expected[i], refValue) {
				return value, nil
			}
		}

		return nil, errors.New("REQUIRED")
	}
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// boolean - checks that validated value is a boolean, 1, 0, "1", "0", "true" or "false", and
// converts it to a boolean.
func boolean(args ...interface{}) Validation {
	noArgs("boolean", args)

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		s, ok := stringForm(value)
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		switch s {
		case "true", "1":
			return true, nil
		case "false", "0":
			return false, nil
		}
		return nil, errors.New("NOT_BOOLEAN")
	}
}

// _any - accepts any value and keeps it unchanged.
func _any(args ...interface{}) Validation {
	noArgs("any", args)

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		return value, nil
	}
}
