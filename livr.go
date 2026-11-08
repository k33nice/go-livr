package livr

import (
	"errors"
	"sync"
	"time"
)

// Version is LIVR current semver
const Version = "2.0.0"

// Dictionary - is dictionary alias.
type Dictionary = map[string]interface{}

// Validation - function in which made all validation checks.
type Validation = func(interface{}, ...interface{}) (interface{}, interface{})

// Builder - common type for building validators.
type Builder = func(...interface{}) Validation

// Validator - Validator object.
type Validator struct {
	once sync.Once

	livrRules         Dictionary
	validators        map[string][]Validation
	validatorBuilders map[string]Builder

	errs      map[string]interface{}
	schemaErr error

	isAutoTrim bool
}

// Options - config for validator instance.
type Options struct {
	LivrRules Dictionary
	AutoTrim  isAutoTrim
	// Now - the clock of the "today", "yesterday" and "tomorrow" bounds of the date rules,
	// time.Now by default (always used in UTC). Tests use it to fix the current date.
	Now func() time.Time
}

var defaultRules map[string]Builder

func init() {
	defaultRules = map[string]Builder{
		// Common related rules.
		"required":       required,
		"not_empty":      notEmpty,
		"not_empty_list": notEmptyList,
		"any_object":     anyObject,
		"required_if":    requiredIf,
		"boolean":        boolean,
		"any":            _any,

		// Text related rules.
		"one_of":         oneOf,
		"eq":             eq,
		"string":         _string,
		"min_length":     minLength,
		"max_length":     maxLength,
		"length_equal":   lengthEqual,
		"length_between": lengthBetween,
		"like":           like,
		"not_one_of":     notOneOf,
		"digits":         digits,
		"starts_with":    startsWith,
		"ends_with":      endsWith,
		"contains":       contains,

		// Rules for real numbers.
		"integer":          integer,
		"positive_integer": positiveInteger,
		"decimal":          decimal,
		"positive_decimal": positiveDecimal,
		"min_number":       minNumber,
		"max_number":       maxNumber,
		"number_between":   numberBetween,
		"max_decimals":     maxDecimals,

		// Misc rules.
		"email":          email,
		"equal_to_field": equalToField,
		"url":            url,
		"iso_date":       isoDate(time.Now),
		"iso_datetime":   isoDatetime(time.Now),
		"iso_time":       isoTime,

		// Formats and checksums.
		"uuid":        uuid,
		"ip":          ip,
		"cidr":        cidr,
		"hostname":    hostname,
		"base64":      base64,
		"hex":         hex,
		"digest":      digest,
		"mongo_id":    mongoID,
		"hex_color":   hexColor,
		"mac":         mac,
		"e164":        e164,
		"semver":      semver,
		"jwt":         jwt,
		"luhn":        luhn,
		"credit_card": creditCard,
		"iban":        iban,

		// Rules for lists.
		"list_length": listLength,
		"unique":      unique,

		// Meta rules.
		"nested_object":             nestedObject,
		"list_of":                   listOf,
		"list_of_objects":           listOfObjects,
		"list_of_different_objects": listOfDifferentObjects,
		"variable_object":           variableObject,
		"or":                        or,
		"when":                      when,
		"values":                    values,
		"keys":                      keys,

		// Modifires - allows to change or sanitize value.
		"default":    _default,
		"trim":       trim,
		"to_lc":      toLc,
		"to_uc":      toUc,
		"remove":     remove,
		"leave_only": leaveOnly,
		"split":      split,
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
		validators:        make(map[string][]Validation),
		isAutoTrim:        at,
	}

	// The default rules are not copied: buildValidator finds them, so that building the
	// validators of nested rules stays cheap.
	if opts.Now != nil {
		v.validatorBuilders["iso_date"] = isoDate(opts.Now)
		v.validatorBuilders["iso_datetime"] = isoDatetime(opts.Now)
	}

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

	return func(args ...interface{}) Validation {
		_, rB := ruleArgs(args)
		validator := fieldValidator(a.Rules, rB)

		return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
			// The rules of the alias see the fields next to the field that uses it.
			res, err := validator.validateField(value, containingObject(builders))
			if err != nil {
				if a.Error != "" {
					return nil, errors.New(a.Error)
				}
				return nil, err
			}
			return res, nil
		}
	}
}

// DefaultRules - return validator default rules.
func (v *Validator) DefaultRules() map[string]Builder {
	return defaultRules
}

// Rules - return validator rules.
func (v *Validator) Rules() map[string][]Validation {
	return v.validators
}

func (v *Validator) registerRules(rules map[string]Builder) {
	for name, r := range rules {
		v.validatorBuilders[name] = r
	}
}

// Validate - validate a data. It panics with a *SchemaError for wrong rules (see Prepare).
func (v *Validator) Validate(data Dictionary) (Dictionary, error) {
	v.prepare()

	if v.isAutoTrim {
		data = autoTrim(data).(Dictionary)
	}

	res, errs := v.validate(data, data)
	v.errs = errs
	if errs != nil {
		return nil, errors.New("validation error")
	}

	return res, nil
}

// Errors - return all validation errors.
func (v *Validator) Errors() Dictionary {
	return v.errs
}

// validate - validate the fields of data. The rules see the fields of params, the containing
// object: data itself, or the object of a field for the rules of an alias, "or" and "when".
func (v *Validator) validate(data, params Dictionary) (Dictionary, Dictionary) {
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
			res, err := validator(val, params)
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
		return nil, errors
	}

	return results, nil
}

// Prepare - build the validator from the LIVR rules. Validate does it on the first call, Prepare
// reports wrong rules before: a rule that is not registered or wrong arguments of a rule give a
// *SchemaError.
func (v *Validator) Prepare() error {
	v.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(*SchemaError)
				if !ok {
					panic(r)
				}
				v.schemaErr = err
			}
		}()

		for field, fieldRules := range v.livrRules {
			if _, ok := fieldRules.([]interface{}); !ok {
				fieldRules = []interface{}{fieldRules}
			}
			var validators []Validation
			for _, rawRule := range fieldRules.([]interface{}) {
				name, args := parseRule(rawRule)
				validators = append(validators, v.buildValidator(name, args))
			}
			v.validators[field] = validators
		}
	})

	return v.schemaErr
}

func (v *Validator) prepare() {
	if err := v.Prepare(); err != nil {
		panic(err)
	}
}

func (v *Validator) buildValidator(name string, args []interface{}) Validation {
	builder, ok := v.validatorBuilders[name]
	if !ok {
		builder, ok = defaultRules[name]
	}
	if !ok && name == "" {
		// A list, an empty object or another value where a rule is expected.
		schemaErrorf("A rule must be a rule name or an object with a rule name and its arguments")
	}
	if !ok {
		schemaErrorf("Rule %s not registered", name)
	}

	args = append(args, v.validatorBuilders)

	return builder(args...)
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
