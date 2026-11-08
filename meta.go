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
		_, rB = ruleArgs(args)
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
		// Arguments after the list of rules are ignored.
		lr = v
		rB = args[len(args)-1]
	} else {
		lr = args[0 : len(args)-1]
		rB = args[len(args)-1]
	}

	validator := fieldValidator(lr, rB.(map[string]Builder))
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
			// Items have no containing object: a reference in their rules is a missing value.
			r, err := validator.validateField(s.Index(i).Interface(), nil)

			if err != nil {
				hasError = true
				errs = append(errs, err)
				results = append(results, nil)
				continue
			}
			results = append(results, r)
			errs = append(errs, nil)
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
		_, rB = ruleArgs(args)
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
		_, rB = ruleArgs(args)
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
		validators = append(validators, fieldValidator(lr, rB))
	}

	return func(val interface{}, builders ...interface{}) (interface{}, interface{}) {
		// No check for an empty value: the sets of rules decide (they may contain "required").
		// The rules see the fields next to the validated field.
		var lastErr interface{}
		for _, validator := range validators {
			r, err := validator.validateField(val, containingObject(builders))

			if err != nil {
				lastErr = err
			} else {
				return r, nil
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
		_, rB = ruleArgs(args)
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

// when - validate the value with rules that depend on another field:
// when(reference, condition, rulesIfTrue[, rulesIfFalse]) or when(reference, cases).
func when(args ...interface{}) Validation {
	args, rB := ruleArgs(args)
	if len(args) < 2 {
		schemaErrorf("Rule when takes a reference and cases, or a reference, a condition and rules")
	}
	path := parseReference(args[0], "when")

	if len(args) == 2 {
		return whenCases(path, args[1], rB)
	}

	condition := branchValidator(args[1], rB)
	ifTrue := branchValidator(args[2], rB)
	var ifFalse *Validator
	if len(args) > 3 {
		ifFalse = branchValidator(args[3], rB)
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		params := containingObject(builders)
		refValue := resolveReference(params, path)

		// The condition rules see the fields next to the validated field, their output is discarded.
		branch := ifFalse
		if !isEmpty(refValue) {
			if _, err := condition.validateField(refValue, params); err == nil {
				branch = ifTrue
			}
		}

		// Without rules if false the value passes unchanged.
		if branch == nil {
			return value, nil
		}
		return branch.validateField(value, params)
	}
}

// whenCases - the switch form of "when": the rules of the case named by the referenced value.
func whenCases(path []string, rawCases interface{}, rB map[string]Builder) Validation {
	cases, ok := rawCases.(Dictionary)
	if !ok || len(cases) == 0 {
		schemaErrorf("Rule when takes an object of cases with at least one case")
	}

	validators := make(map[string]*Validator, len(cases))
	for name, rules := range cases {
		validators[name] = branchValidator(rules, rB)
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		params := containingObject(builders)
		refValue := resolveReference(params, path)

		var branch *Validator
		if !isEmpty(refValue) {
			if name, ok := stringForm(refValue); ok {
				branch = validators[name]
			}
		}

		// An empty or unknown referenced value chooses no case.
		if branch == nil {
			if isEmpty(value) {
				return value, nil
			}
			return nil, errors.New("NOT_ALLOWED_VALUE")
		}
		return branch.validateField(value, params)
	}
}

// branchValidator - a validator of the rules of a branch of "when", not an empty list of rules.
func branchValidator(rules interface{}, rB map[string]Builder) *Validator {
	if list, ok := rules.([]interface{}); ok && len(list) == 0 {
		schemaErrorf(`Rule when got an empty list of rules, use "any" for no rules`)
	}
	return fieldValidator(rules, rB)
}

// rulesArg - the rules of "values" and "keys", given like the rules of list_of:
// values("required", "string") or values(["required", "string"]).
func rulesArg(name string, args []interface{}) ([]interface{}, map[string]Builder) {
	args, rB := ruleArgs(args)
	if len(args) > 0 {
		if list, ok := args[0].([]interface{}); ok {
			args = list
		}
	}
	if len(args) == 0 {
		schemaErrorf("Rule %s needs rules", name)
	}
	return args, rB
}

// values - check that validated value is an object and validate its every value.
func values(args ...interface{}) Validation {
	validator := fieldValidator(rulesArg("values", args))

	return func(object interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(object) {
			return object, nil
		}

		obj, ok := object.(Dictionary)
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		results := make(Dictionary, len(obj))
		errs := make(Dictionary)
		for key, value := range obj {
			// Values have no containing object: a reference in their rules is a missing value.
			r, err := validator.validateField(value, nil)
			if err != nil {
				errs[key] = err
			} else {
				results[key] = r
			}
		}

		if len(errs) > 0 {
			return nil, errs
		}
		return results, nil
	}
}

// keys - check that validated value is an object and check its every key.
func keys(args ...interface{}) Validation {
	validator := fieldValidator(rulesArg("keys", args))

	return func(object interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(object) {
			return object, nil
		}

		obj, ok := object.(Dictionary)
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		errs := make(Dictionary)
		for key := range obj {
			// The rules only check the key, their output is ignored.
			if _, err := validator.validateField(key, nil); err != nil {
				errs[key] = errors.New("WRONG_KEY")
			}
		}

		if len(errs) > 0 {
			return nil, errs
		}
		return object, nil
	}
}
