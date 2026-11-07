package test

import (
	"testing"
	"time"

	"github.com/k33nice/go-livr"
)

func TestPrepare(t *testing.T) {
	v := livr.New(&livr.Options{LivrRules: livr.Dictionary{"id": livr.Dictionary{"uuid": 4}}})
	err := v.Prepare()
	if _, ok := err.(*livr.SchemaError); !ok {
		t.Fatalf("got %v, want a *livr.SchemaError", err)
	}

	v = livr.New(&livr.Options{LivrRules: livr.Dictionary{"id": "no_such_rule"}})
	if err := v.Prepare(); err == nil || err.Error() != "Rule no_such_rule not registered" {
		t.Errorf("got %v, want the error of an unknown rule", err)
	}

	v = livr.New(&livr.Options{LivrRules: livr.Dictionary{"id": "uuid"}})
	if err := v.Prepare(); err != nil {
		t.Errorf("got %v, want no error", err)
	}
}

func TestNow(t *testing.T) {
	now := func() time.Time { return time.Date(2026, 10, 18, 23, 30, 0, 0, time.FixedZone("", -3600)) }
	rules := livr.Dictionary{
		"date":     livr.Dictionary{"iso_date": livr.Dictionary{"max": "today"}},
		"datetime": livr.Dictionary{"list_of": livr.Dictionary{"iso_datetime": livr.Dictionary{"max": "today"}}},
	}

	// The current date is in UTC: 2026-10-19.
	v := livr.New(&livr.Options{LivrRules: rules, Now: now})
	input := livr.Dictionary{"date": "2026-10-19", "datetime": []interface{}{"2026-10-19T23:59:59.999Z"}}
	if _, err := v.Validate(input); err != nil {
		t.Errorf("got errors %v", indirectErrors(v.Errors()))
	}

	input = livr.Dictionary{"date": "2026-10-20", "datetime": []interface{}{"2026-10-20T00:00Z"}}
	want := `{"date":"DATE_TOO_LATE","datetime":["DATETIME_TOO_LATE"]}`
	if _, err := v.Validate(input); err == nil || toJSON(indirectErrors(v.Errors())) != want {
		t.Errorf("got errors %s, want %s", toJSON(indirectErrors(v.Errors())), want)
	}
}

func TestIntegerArguments(t *testing.T) {
	// Rules written in Go may have integers as arguments.
	rules := livr.Dictionary{
		"amount": livr.Dictionary{"max_decimals": 2},
		"tags":   livr.Dictionary{"list_length": []interface{}{1, 2}},
		"mac":    livr.Dictionary{"mac": 64},
	}
	input := livr.Dictionary{"amount": 1.255, "tags": []interface{}{}, "mac": "00:1a:2b:3c:4d:5e"}
	want := `{"errors":{"amount":"TOO_MANY_DECIMALS","mac":"WRONG_MAC","tags":"TOO_FEW_ITEMS"}}`
	if got := validateJSON(rules, input); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestSchemaErrorsInRuleLists(t *testing.T) {
	for _, rules := range []livr.Dictionary{
		{"a": []interface{}{"required", []interface{}{}}},
		{"a": livr.Dictionary{"when": []interface{}{"b", "required", []interface{}{"required", []interface{}{}}}}},
		{"a": livr.Dictionary{"values": []interface{}{"required", []interface{}{}}}},
		{"a": livr.Dictionary{"when": []interface{}{1.0, "required", "required"}}},
		{"a": livr.Dictionary{"list_length": []interface{}{3, "x"}}},
		{"a": livr.Dictionary{"max_decimals": " 2"}},
		{"a": livr.Dictionary{"max_decimals": "1e1"}},
		{"a": livr.Dictionary{"mac": "0x40"}},
	} {
		if err := livr.New(&livr.Options{LivrRules: rules}).Prepare(); err == nil {
			t.Errorf("%s: the validator was built, must fail", toJSON(rules))
		}
	}

	// Ignored arguments are not checked; a number in the decimal grammar may be a string.
	for _, rules := range []livr.Dictionary{
		{"a": livr.Dictionary{"split": []interface{}{",", livr.Dictionary{}}}},
		{"a": livr.Dictionary{"list_length": []interface{}{livr.Dictionary{"max": 2.0}, "x"}}},
		{"a": livr.Dictionary{"max_decimals": "02"}},
		{"a": livr.Dictionary{"max_decimals": "2.0"}},
		{"a": livr.Dictionary{"iso_date": livr.Dictionary{"min": "2099-01-01", "max": "today"}}},
	} {
		if err := livr.New(&livr.Options{LivrRules: rules}).Prepare(); err != nil {
			t.Errorf("%s: got %v, want no error", toJSON(rules), err)
		}
	}
}

func TestBooleansInEqAndOneOf(t *testing.T) {
	// "true" and "false" match a boolean, and one_of keeps the strconv.ParseBool forms of LIVR 2.0.
	rules := livr.Dictionary{
		"eq":         livr.Dictionary{"eq": true},
		"eq_string":  livr.Dictionary{"eq": "true"},
		"one_of":     livr.Dictionary{"one_of": []interface{}{[]interface{}{true}}},
		"one_of_f":   livr.Dictionary{"one_of": []interface{}{[]interface{}{false}}},
		"one_of_str": livr.Dictionary{"one_of": []interface{}{[]interface{}{"true"}}},
	}
	input := livr.Dictionary{"eq": "true", "eq_string": true, "one_of": "true", "one_of_f": "false", "one_of_str": true}
	want := `{"eq":"true","eq_string":true,"one_of":true,"one_of_f":false,"one_of_str":"true"}`
	if got := validateJSON(rules, input); got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	for _, s := range []string{"1", "t", "T", "TRUE", "true", "True"} {
		input := livr.Dictionary{"one_of": s, "one_of_f": s}
		want := `{"errors":{"one_of_f":"NOT_ALLOWED_VALUE"}}`
		if got := validateJSON(rules, input); got != want {
			t.Errorf("%q: got %s, want %s", s, got, want)
		}
	}
	for _, s := range []string{"0", "f", "F", "FALSE", "false", "False"} {
		input := livr.Dictionary{"one_of": s, "one_of_f": s}
		want := `{"errors":{"one_of":"NOT_ALLOWED_VALUE"}}`
		if got := validateJSON(rules, input); got != want {
			t.Errorf("%q: got %s, want %s", s, got, want)
		}
	}

	// Other strings and numbers are not booleans; eq did not take the ParseBool forms in 2.0.
	input = livr.Dictionary{"eq": "1", "one_of": "yes", "one_of_f": 0.0, "one_of_str": false}
	want = `{"errors":{"eq":"NOT_ALLOWED_VALUE","one_of":"NOT_ALLOWED_VALUE","one_of_f":"NOT_ALLOWED_VALUE","one_of_str":"NOT_ALLOWED_VALUE"}}`
	if got := validateJSON(rules, input); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNullArgument(t *testing.T) {
	// LIVR 2.0 ignored the arguments of iso_date, so {"iso_date": null} is the plain iso_date.
	rules := livr.Dictionary{"date": livr.Dictionary{"iso_date": nil}}
	input := livr.Dictionary{"date": "2026-02-29"}
	want := `{"errors":{"date":"WRONG_DATE"}}`
	if got := validateJSON(rules, input); got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	// The LIVR 2.0 rules and custom rules get null as one argument.
	var customArgs []interface{}
	v := livr.New(&livr.Options{LivrRules: livr.Dictionary{
		"name":  livr.Dictionary{"default": nil},
		"other": livr.Dictionary{"null_argument_probe": nil},
	}})
	v.RegisterDefaultRules(map[string]livr.Builder{
		"null_argument_probe": func(args ...interface{}) livr.Validation {
			customArgs = args[:len(args)-1]
			return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
				return value, nil
			}
		},
	})
	output, err := v.Validate(livr.Dictionary{"name": ""})
	if err != nil {
		t.Fatalf("got errors %v", indirectErrors(v.Errors()))
	}
	// A rule that returns null keeps the value, so {"default": null} keeps "" as in LIVR 2.0.
	if value, ok := output["name"]; !ok || value != "" {
		t.Errorf("default: got %v, want the value unchanged", output)
	}
	if len(customArgs) != 1 || customArgs[0] != nil {
		t.Errorf("custom rule: got arguments %v, want [null]", customArgs)
	}

	// {"rule": null} is one null argument, which no LIVR 2.5 rule takes. A null bound and a null
	// inside a list of arguments are wrong values too.
	for _, rules := range []livr.Dictionary{
		{"a": livr.Dictionary{"boolean": nil}},
		{"a": livr.Dictionary{"any": nil}},
		{"a": livr.Dictionary{"uuid": nil}},
		{"a": livr.Dictionary{"uuid": []interface{}{nil}}},
		{"a": livr.Dictionary{"ip": nil}},
		{"a": livr.Dictionary{"base64": nil}},
		{"a": livr.Dictionary{"mac": nil}},
		{"a": livr.Dictionary{"unique": nil}},
		{"a": livr.Dictionary{"iso_datetime": nil}},
		{"a": livr.Dictionary{"iso_time": nil}},
		{"a": livr.Dictionary{"values": nil}},
		{"a": livr.Dictionary{"list_length": livr.Dictionary{"min": nil}}},
		{"a": livr.Dictionary{"list_length": nil}},
		{"a": livr.Dictionary{"iso_date": livr.Dictionary{"max": nil}}},
		{"a": livr.Dictionary{"not_one_of": []interface{}{nil}}},
		{"a": livr.Dictionary{"not_one_of": []interface{}{"x", nil}}},
		{"a": livr.Dictionary{"required_if": livr.Dictionary{"type": nil}}},
		{"a": livr.Dictionary{"split": nil}},
	} {
		if err := livr.New(&livr.Options{LivrRules: rules}).Prepare(); err == nil {
			t.Errorf("%s: the validator was built, must fail", toJSON(rules))
		}
	}
}

func TestUniqueKeyValues(t *testing.T) {
	// A key value that is an object or a list is a wrong structure, also next to an empty key.
	rules := livr.Dictionary{"items": livr.Dictionary{"unique": []interface{}{"a", "b"}}}
	for _, item := range []livr.Dictionary{
		{"a": livr.Dictionary{}, "b": ""},
		{"a": nil, "b": []interface{}{1.0}},
		{"b": livr.Dictionary{"c": 1.0}},
	} {
		input := livr.Dictionary{"items": []interface{}{item}}
		want := `{"errors":{"items":"FORMAT_ERROR"}}`
		if got := validateJSON(rules, input); got != want {
			t.Errorf("%s: got %s, want %s", toJSON(item), got, want)
		}
	}
}
