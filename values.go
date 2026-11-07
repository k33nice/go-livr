package livr

import (
	"fmt"
	"math"
	"strconv"
)

// SchemaError - an error in the LIVR rules: a rule that is not registered or an argument that
// does not match the signature of a rule. Prepare returns it, Validate panics with it.
type SchemaError struct {
	Message string
}

func (e *SchemaError) Error() string {
	return e.Message
}

// schemaErrorf - stop building the validator with a SchemaError.
func schemaErrorf(format string, a ...interface{}) {
	panic(&SchemaError{Message: fmt.Sprintf(format, a...)})
}

// ruleArgs - split the arguments of a rule builder into the arguments of the rule and the rule
// builders, which are always the last argument. {"rule": null} is one null argument.
func ruleArgs(args []interface{}) ([]interface{}, map[string]Builder) {
	if len(args) == 0 {
		return args, nil
	}
	rB, _ := args[len(args)-1].(map[string]Builder)
	return args[:len(args)-1], rB
}

// noArgs - a schema error for a rule without arguments that gets any. args are the arguments of
// the rule builder, with the rule builders last.
func noArgs(name string, args []interface{}) {
	if args, _ = ruleArgs(args); len(args) > 0 {
		schemaErrorf("Rule %s takes no arguments", name)
	}
}

// isEmpty - a missing value, null or "".
func isEmpty(value interface{}) bool {
	return value == nil || value == ""
}

// number - the value of a number: float64 from JSON or an integer from Go code.
func number(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, !math.IsInf(v, 0) && !math.IsNaN(v)
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	}
	return 0, false
}

// formatNumber - the string form of a number: the shortest decimal that reads back as the same
// number, without an exponent ("1", "0.5", "-12.25").
func formatNumber(f float64) string {
	if f == 0 {
		return "0" // and not "-0"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// stringForm - the string form of a primitive value (a string, a number or a boolean);
// false for other values.
func stringForm(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case bool:
		return strconv.FormatBool(v), true
	}
	if f, ok := number(value); ok {
		return formatNumber(f), true
	}
	return "", false
}

// stringArg - a string argument: a string, or a number in its string form.
func stringArg(arg interface{}) (string, bool) {
	if _, ok := arg.(bool); ok {
		return "", false
	}
	return stringForm(arg)
}

// numberArg - a number argument: a number, or a string with a number in the grammar of "decimal".
func numberArg(arg interface{}) (float64, bool) {
	if s, ok := arg.(string); ok {
		if countDecimals(s) < 0 {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil
	}
	return number(arg)
}

// countArg - a non-negative integer argument (a count or a length).
func countArg(arg interface{}) (int, bool) {
	f, ok := numberArg(arg)
	if !ok || f < 0 || f != math.Trunc(f) || f > math.MaxInt32 {
		return 0, false
	}
	return int(f), true
}

// countDecimals - the number of digits after the decimal point of a string in the grammar of the
// "decimal" rule, -?[0-9]+(\.[0-9]+)?, without trailing zeros: "1.230" has 2, "5.000" has 0.
// -1 when the string is not such a number.
func countDecimals(s string) int {
	i := 0
	if i < len(s) && s[i] == '-' {
		i++
	}
	start := i
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	if i == start {
		return -1
	}
	if i == len(s) {
		return 0
	}
	if s[i] != '.' || i+1 == len(s) {
		return -1
	}
	fractionStart := i + 1
	significantEnd := fractionStart
	for i = fractionStart; i < len(s); i++ {
		if !isDigit(s[i]) {
			return -1
		}
		if s[i] != '0' {
			significantEnd = i + 1
		}
	}
	return significantEnd - fractionStart
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// fieldValidator - a validator of a single value with rules in any form that a field accepts,
// for the meta rules.
func fieldValidator(rules interface{}, rB map[string]Builder) *Validator {
	v := New(&Options{LivrRules: Dictionary{"field": rules}})
	v.registerRules(rB)
	v.prepare()
	return v
}

// validateField - validate value with the rules of fieldValidator. The rules see the fields of
// params, the containing object. It returns the output (nil to keep the value) and the error.
func (v *Validator) validateField(value interface{}, params Dictionary) (interface{}, interface{}) {
	res, errs := v.validate(Dictionary{"field": value}, params)
	if errs != nil {
		return nil, errs["field"]
	}
	return res["field"], nil
}

// containingObject - the containing object that the rules get as their second argument.
func containingObject(builders []interface{}) Dictionary {
	if len(builders) > 0 {
		if params, ok := builders[0].(Dictionary); ok {
			return params
		}
	}
	return nil
}
