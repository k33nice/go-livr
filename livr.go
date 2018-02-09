package livr

import (
	"errors"
	"log"
	"sync"
)

// Version is LIVR current semver
const Version = "2.0.0-alpha"

// Dictionary - is dictionary alias.
type Dictionary = map[string]interface{}

// Builder - common type for building validators.
type Builder = func(...interface{}) func(...interface{}) (interface{}, interface{})

// Validator - Validator object.
type Validator struct {
	once sync.Once

	livrRules         Dictionary
	validators        map[string][]func(...interface{}) (interface{}, interface{})
	validatorBuilders map[string]Builder

	errs map[string]interface{}

	isAutoTrim bool
}

// Options - config for validator instance.
type Options struct {
	LivrRules Dictionary
	AutoTrim  isAutoTrim
}

var defaultRules map[string]Builder

func init() {
	defaultRules = map[string]Builder{
		// Common related rules.
		"required":       required,
		"not_empty":      notEmpty,
		"not_empty_list": notEmptyList,
		"any_object":     anyObject,

		// Text related rules.
		"one_of":         oneOf,
		"eq":             eq,
		"string":         _string,
		"min_length":     minLength,
		"max_length":     maxLength,
		"length_equal":   lengthEqual,
		"length_between": lengthBetween,
		"like":           like,

		// Rules for real numbers.
		"integer":          integer,
		"positive_integer": positiveInteger,
		"decimal":          decimal,
		"positive_decimal": positiveDecimal,
		"min_number":       minNumber,
		"max_number":       maxNumber,
		"number_between":   numberBetween,

		// Misc rules.
		"email":          email,
		"equal_to_field": equalToField,
		"url":            url,
		"iso_date":       isoDate,

		// Meta rules.
		"nested_object":             nestedObject,
		"list_of":                   listOf,
		"list_of_objects":           listOfObjects,
		"list_of_different_objects": listOfDifferentObjects,
		"variable_object":           variableObject,
		"or":                        or,

		// Modifires - allows to change or sanitize value.
		"default":    _default,
		"trim":       trim,
		"to_lc":      toLc,
		"to_uc":      toUc,
		"remove":     remove,
		"leave_only": leaveOnly,
	}
}

// New - return new instance of Validator.
func New(opts *Options) *Validator {
	at := defaultAutoTrim
	if opts.AutoTrim != Nil {
		at = opts.AutoTrim.Bool()
	}
	v := &Validator{
		livrRules:         opts.LivrRules,
		validatorBuilders: make(map[string]Builder),
		validators:        make(map[string][]func(...interface{}) (interface{}, interface{})),
		isAutoTrim:        at,
	}

	v.registerRules(defaultRules)

	return v
}

// RegisterDefaultRules - register custom user rules.
func (v *Validator) RegisterDefaultRules(rules map[string]Builder) {
	for name, rule := range rules {
		defaultRules[name] = rule
	}
}

// Alias - for defining user aliases.
type Alias struct {
	Name  string      `json:"name"`
	Error string      `json:"error"`
	Rules interface{} `json:"rules"`
}

// RegisterAliasedDefaultRule - make alias to default rule.
func (v *Validator) RegisterAliasedDefaultRule(a Alias) {
	defaultRules[a.Name] = v.buildAliasedRule(a)
}

// RegisterAliasedRule - make alias to default rule.
func (v *Validator) RegisterAliasedRule(a Alias) {
	v.validatorBuilders[a.Name] = v.buildAliasedRule(a)
}

func (v *Validator) buildAliasedRule(a Alias) Builder {
	if a.Name == "" {
		panic("Alias name required")
	}

	if a.Rules == nil {
		panic("Alias rules required")
	}

	return func(args ...interface{}) func(...interface{}) (interface{}, interface{}) {
		validator := New(&Options{LivrRules: Dictionary{"value": a.Rules}})
		validator.registerRules(args[0].(map[string]Builder))
		validator.prepare()

		return func(builders ...interface{}) (interface{}, interface{}) {
			var value interface{}
			if len(builders) > 0 {
				value = builders[0]
			}
			res, err := validator.Validate(Dictionary{"value": value})
			if err != nil {
				errs := validator.Errors()
				if a.Error != "" {
					return nil, errors.New(a.Error)
				}
				if err, ok := errs["value"]; ok {
					return nil, err
				}

				return nil, validator.Errors()["value"]
			}
			if out, ok := res["value"]; ok {
				return out, nil
			}
			return nil, nil
		}
	}
}

// DefaultRules - return validator default rules.
func (v *Validator) DefaultRules() map[string]Builder {
	return defaultRules
}

// Rules - return validator rules.
func (v *Validator) Rules() map[string][]func(...interface{}) (interface{}, interface{}) {
	return v.validators
}

func (v *Validator) registerRules(rules map[string]Builder) {
	for name, r := range rules {
		v.validatorBuilders[name] = r
	}
}

// Validate - validate a data.
func (v *Validator) Validate(data Dictionary) (Dictionary, error) {
	v.prepare()

	res := v.validate(data)
	if res == nil {
		return nil, errors.New("validation error")
	}

	return res, nil
}

// Errors - return all validation errors.
func (v *Validator) Errors() Dictionary {
	return v.errs
}

func (v *Validator) validate(data Dictionary) Dictionary {
	results := make(Dictionary)
	errors := make(Dictionary)

	for fName, validators := range v.validators {
		if len(validators) == 0 {
			continue
		}

		var val interface{}
		if _, ok := data[fName]; ok {
			val = data[fName]
		}

		for _, validator := range validators {
			if v, ok := results[fName]; ok {
				val = v
			}
			res, err := validator(val, data)
			if err != nil {
				errors[fName] = err
				break
			} else if res != nil {
				results[fName] = res
			} else if _, ok := data[fName]; ok {
				results[fName] = val
			}
		}
	}

	if len(errors) > 0 {
		v.errs = errors
		return nil
	}

	v.errs = nil

	return results
}

func (v *Validator) prepare() {
	v.once.Do(func() {
		for field, fieldRules := range v.livrRules {
			if _, ok := fieldRules.([]interface{}); !ok {
				fieldRules = []interface{}{fieldRules}
			}
			var validators []func(...interface{}) (interface{}, interface{})
			for _, rawRule := range fieldRules.([]interface{}) {
				name, args := parseRule(rawRule)
				validators = append(validators, v.buildValidator(name, args))
			}
			v.validators[field] = validators
		}
	})
}

func (v *Validator) buildValidator(name string, args []interface{}) func(...interface{}) (interface{}, interface{}) {
	if _, ok := v.validatorBuilders[name]; !ok {
		log.Panicf("Rule %s not registerd", name)
	}

	args = append(args, v.validatorBuilders)

	return v.validatorBuilders[name](args...)
}

func parseRule(lr interface{}) (string, []interface{}) {
	switch rule := lr.(type) {
	case map[string]interface{}:
		for name, args := range rule {
			if args, ok := args.([]interface{}); ok {
				return name, args
			}

			return name, []interface{}{args}
		}

		return "", []interface{}{}
	case string:
		return rule, []interface{}{}
	default:
		return "", []interface{}{}
	}
}
