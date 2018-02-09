package livr

import (
	"errors"
	"reflect"
)

// NestedObject - check that validated value is object.
func NestedObject(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
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

	return func(builders ...interface{}) (interface{}, interface{}) {
		var nestedObject interface{}
		if len(builders) > 0 {
			nestedObject = builders[0]
		}

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

// ListOf - check that validated value is list of some objects.
func ListOf(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var firstArg interface{}
	if len(args) > 0 {
		firstArg = args[0]
	}

	var lr, rB interface{}
	if v, ok := firstArg.([]interface{}); ok {
		lr = v
		if len(args) > 1 {
			rB = args[1]
		}
	} else {
		lr = args[0 : len(args)-1]
		rB = args[len(args)-1]
	}

	validator := New(&Options{LivrRules: Dictionary{"field": lr}})
	validator.prepare()
	validator.registerRules(rB.(map[string]Builder))
	return func(builders ...interface{}) (interface{}, interface{}) {
		var values interface{}
		if len(builders) > 0 {
			values = builders[0]
		}

		if values == nil || values == "" {
			return nil, nil
		}

		rt := reflect.TypeOf(values)

		if rt.Kind() != reflect.Slice {
			return nil, errors.New("FORMAT_ERROR")
		}

		s := reflect.ValueOf(values)
		var results, errs []interface{}
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

// ListOfObjects - check that validated value is list of some objects.
func ListOfObjects(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
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
	validator.prepare()
	validator.registerRules(rB)
	return func(builders ...interface{}) (interface{}, interface{}) {
		var objects interface{}
		if len(builders) > 0 {
			objects = builders[0]
		}

		if objects == nil || objects == "" {
			return nil, nil
		}

		rt := reflect.TypeOf(objects)

		if rt.Kind() != reflect.Slice {
			return nil, errors.New("FORMAT_ERROR")
		}

		s := reflect.ValueOf(objects)

		var results, errs []interface{}
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

// ListOfDifferentObjects - checks that validated value is one of specified objects.
func ListOfDifferentObjects(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
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
		validator.prepare()
		validator.registerRules(rB)
		validators[selVal] = validator
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var objects []interface{}
		if len(builders) > 0 {
			if v, ok := builders[0].([]interface{}); ok {
				objects = v
			}
		}

		var results, errs []interface{}
		var hasError bool
		for _, object := range objects {
			if _, ok := object.(Dictionary); !ok {
				errs = append(errs, errors.New("FORMAT_ERROR"))
				continue
			}
			if _, ok := object.(Dictionary)[selField]; !ok {
				errs = append(errs, errors.New("FORMAT_ERROR"))
				continue
			}
			// TODO: check it.
			if v, ok := validators[object.(Dictionary)[selField].(string)]; !ok || v == nil {
				errs = append(errs, errors.New("FORMAT_ERROR"))
				continue
			}

			v := validators[object.(Dictionary)[selField].(string)]
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

// Or - check that validated value is one of specified.
func Or(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
	var firstArg interface{}
	if len(args) > 0 {
		firstArg = args[0]
	}

	var lrs []interface{}
	var rB map[string]Builder
	if v, ok := firstArg.([]interface{}); ok {
		lrs = v
		if v, ok := args[2].(map[string]Builder); ok {
			rB = v
		}
		if len(args) > 1 {
			if v, ok := args[1].(map[string]Builder); ok {
				rB = v
			}
		}
	} else {
		lrs = args[0 : len(args)-1]
		if v, ok := args[len(args)-1].(map[string]Builder); ok {
			rB = v
		}
	}

	var validators []*Validator

	for _, lr := range lrs {
		validator := New(&Options{LivrRules: Dictionary{"field": lr}})
		validator.prepare()
		validator.registerRules(rB)
		validators = append(validators, validator)
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var val interface{}
		if len(builders) > 1 {
			val = builders[0]
		}

		if val == nil || val == "" {
			return nil, nil
		}

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

// VariableObject - check that validated value is one of specified depends on some inner value.
func VariableObject(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
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
		validator.prepare()
		validator.registerRules(rB)
		validators[selVal] = validator
	}

	return func(builders ...interface{}) (interface{}, interface{}) {
		var object interface{}
		if len(builders) > 1 {
			object = builders[0]
		}

		if _, ok := object.(Dictionary); !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		if _, ok := object.(Dictionary)[selField]; !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		// TODO: check it.
		if v, ok := validators[object.(Dictionary)[selField].(string)]; !ok || v == nil {
			return nil, errors.New("FORMAT_ERROR")
		}

		v := validators[object.(Dictionary)[selField].(string)]

		r, err := v.Validate(object.(Dictionary))
		if err != nil {
			return nil, v.errs
		}
		return r, nil
	}
}
