package livr

import (
	"errors"
	"reflect"
	"strconv"
)

// nestedObject - check that validated value is object.
func nestedObject(args ...interface{}) Validation {
	var lr Dictionary
	var rB map[string]Builder
	if len(args) > 1 {
		if v, ok := args[0].(Dictionary); ok {
			lr = v
		}
		if v, ok := args[1].(map[string]Builder); ok {
			rB = v
		}

	}

	validator := New(&Options{LivrRules: lr})
	validator.registerRules(rB)
	validator.prepare()

	return func(nestedObject interface{}, builders ...interface{}) (interface{}, interface{}) {
		if nestedObject == nil || nestedObject == "" {
			return nil, nil
		}

		if _, ok := nestedObject.(Dictionary); !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		r, err := validator.Validate(nestedObject.(Dictionary))

		if err != nil {
			return nil, validator.errs
		}
		return r, nil
	}
}

// listOf - check that validated value is list of some objects.
func listOf(args ...interface{}) Validation {
	fa := firstArg(args...)

	var lr, rB interface{}
	if v, ok := fa.([]interface{}); ok {
		lr = v
		if len(args) > 1 {
			rB = args[1]
		}
	} else {
		lr = args[0 : len(args)-1]
		rB = args[len(args)-1]
	}

	validator := New(&Options{LivrRules: Dictionary{"field": lr}})
	validator.registerRules(rB.(map[string]Builder))
	validator.prepare()
	return func(values interface{}, builders ...interface{}) (interface{}, interface{}) {
		if values == nil || values == "" {
			return nil, nil
		}

		rt := reflect.TypeOf(values)

		if rt.Kind() != reflect.Slice {
			return nil, errors.New("FORMAT_ERROR")
		}

		s := reflect.ValueOf(values)
		results := make([]interface{}, 0, s.Len())
		var errs []interface{}
		var hasError bool
		for i := 0; i < s.Len(); i++ {
			r, err := validator.Validate(Dictionary{"field": s.Index(i).Interface()})

			if err != nil {
				hasError = true
				errs = append(errs, validator.errs["field"])
				results = append(results, nil)
				continue
			} else {
				if res, ok := r["field"]; ok {
					results = append(results, res)
					errs = append(errs, nil)
					continue
				}
				results = append(results, nil)
				errs = append(errs, nil)
			}
		}

		if hasError {
			return nil, errs
		}
		return results, nil
	}
}

// listOfObjects - check that validated value is list of some objects.
func listOfObjects(args ...interface{}) Validation {
	var lr Dictionary
	var rB map[string]Builder
	if len(args) > 1 {
		if v, ok := args[0].(Dictionary); ok {
			lr = v
		}
		if v, ok := args[1].(map[string]Builder); ok {
			rB = v
		}
	}

	validator := New(&Options{LivrRules: lr})
	validator.registerRules(rB)
	validator.prepare()
	return func(objects interface{}, builders ...interface{}) (interface{}, interface{}) {
		if objects == nil || objects == "" {
			return objects, nil
		}

		rt := reflect.TypeOf(objects)

		if rt.Kind() != reflect.Slice {
			return nil, errors.New("FORMAT_ERROR")
		}

		s := reflect.ValueOf(objects)

		results := make([]interface{}, 0, s.Len())
		var errs []interface{}
		var hasError bool
		for i := 0; i < s.Len(); i++ {

			if _, ok := s.Index(i).Interface().(Dictionary); !ok {
				hasError = true
				errs = append(errs, errors.New("FORMAT_ERROR"))
				results = append(results, nil)
				continue
			}

			r, err := validator.Validate(s.Index(i).Interface().(Dictionary))

			if err != nil {
				hasError = true
				errs = append(errs, validator.errs)
				results = append(results, nil)
			} else {
				results = append(results, r)
				errs = append(errs, nil)
			}
		}
		if hasError {
			return nil, errs
		}

		return results, nil
	}
}

// listOfDifferentObjects - checks that validated value is one of specified objects.
func listOfDifferentObjects(args ...interface{}) Validation {
	var validators = make(map[string]*Validator)

	var selField string
	var lrs Dictionary

	var rB map[string]Builder
	if len(args) > 2 {
		if v, ok := args[0].(string); ok {
			selField = v
		}
		if v, ok := args[1].(Dictionary); ok {
			lrs = v
		}
		if v, ok := args[2].(map[string]Builder); ok {
			rB = v
		}
	}

	for selVal, lr := range lrs {
		rules, ok := lr.(Dictionary)
		if !ok {
			continue
		}
		validator := New(&Options{LivrRules: rules})
		validator.registerRules(rB)
		validator.prepare()
		validators[selVal] = validator
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return nil, nil
		}

		objects, ok := value.([]interface{})
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		results := make([]interface{}, 0, len(objects))
		var errs []interface{}
		var hasError bool
		for _, object := range objects {
			v := selectValidator(validators, object, selField)
			if v == nil {
				hasError = true
				errs = append(errs, errors.New("FORMAT_ERROR"))
				continue
			}

			r, err := v.Validate(object.(Dictionary))

			if err == nil {
				results = append(results, r)
				errs = append(errs, nil)
			} else {
				results = append(results, nil)
				errs = append(errs, v.errs)
				hasError = true
			}
		}

		if hasError {
			return nil, errs
		}

		return results, nil
	}
}

// or - check that validated value is one of specified.
func or(args ...interface{}) Validation {
	// Every argument but the last (the rule builders) is a set of rules.
	lrs := args[0 : len(args)-1]
	var rB map[string]Builder
	if v, ok := args[len(args)-1].(map[string]Builder); ok {
		rB = v
	}

	var validators []*Validator

	for _, lr := range lrs {
		validator := New(&Options{LivrRules: Dictionary{"field": lr}})
		validator.registerRules(rB)
		validator.prepare()
		validators = append(validators, validator)
	}

	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
		// No check for an empty value: the sets of rules decide (they may contain "required").
		var lastErr interface{}
		for _, validator := range validators {
			r, err := validator.Validate(Dictionary{"field": val})

			if err != nil && r == nil {
				lastErr = validator.Errors()["field"]
			} else {
				return r["field"], nil
			}
		}

		if lastErr != nil {
			return nil, lastErr
		}

		return nil, nil
	}
}

// variableObject - check that validated value is one of specified depends on some inner value.
func variableObject(args ...interface{}) Validation {
	var validators = make(map[string]*Validator)

	var selField string
	var lrs Dictionary
	var rB map[string]Builder
	if len(args) > 2 {
		if v, ok := args[0].(string); ok {
			selField = v
		}
		if v, ok := args[1].(Dictionary); ok {
			lrs = v
		}
		if v, ok := args[2].(map[string]Builder); ok {
			rB = v
		}
	}

	for selVal, lr := range lrs {
		rules, ok := lr.(Dictionary)
		if !ok {
			continue
		}
		validator := New(&Options{LivrRules: rules})
		validator.registerRules(rB)
		validator.prepare()
		validators[selVal] = validator
	}

	return func(object interface{}, builders ...interface{}) (interface{}, interface{}) {
		if object == nil || object == "" {
			return nil, nil
		}

		v := selectValidator(validators, object, selField)
		if v == nil {
			return nil, errors.New("FORMAT_ERROR")
		}

		r, err := v.Validate(object.(Dictionary))
		if err != nil {
			return nil, v.errs
		}
		return r, nil
	}
}

// selectValidator - return the validator chosen by the selector field of object,
// or nil if object is not an object or the selector value is not a non-empty string,
// a non-zero number or true listed in the rules.
func selectValidator(validators map[string]*Validator, object interface{}, selField string) *Validator {
	obj, ok := object.(Dictionary)
	if !ok {
		return nil
	}

	var selVal string
	switch v := obj[selField].(type) {
	case string:
		if v == "" {
			return nil
		}
		selVal = v
	case float64:
		if v == 0 {
			return nil
		}
		selVal = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if !v {
			return nil
		}
		selVal = "true"
	default:
		return nil
	}

	return validators[selVal]
}
