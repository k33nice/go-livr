## Validator LIVR

[![Go](https://github.com/k33nice/go-livr/actions/workflows/go.yml/badge.svg)](https://github.com/k33nice/go-livr/actions/workflows/go.yml)
[![Documentation](https://godoc.org/github.com/k33nice/go-livr?status.svg)](https://godoc.org/github.com/k33nice/go-livr)
[![release](https://img.shields.io/badge/release-v2.0.0-blue.svg)](https://github.com/k33nice/go-livr/releases/tag/v2.0.0)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fk33nice%2Fgo-livr.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fk33nice%2Fgo-livr?ref=badge_shield)

Validator LIVR - Lightweight validator supporting Language Independent Validation Rules Specification (LIVR).

### Introduction

One of the biggest advantages is language independent. There are implementations on other languages so you can use your
validation rules in applications on different programming languages.
For example, if you write web application with back-end on Go, you can use your validation rules in front-end (JS)
and back-end (Go).
By reusing your rules, you solve one the most tricky problem of inconsitent validation of the same data.

## USAGE
1. Download and install.
```sh
go get github.com/k33nice/go-livr
```

2. Example.
```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/k33nice/go-livr"
)

func main() {
	var jsonRules = `{
		"name":      "required",
		"email":     ["required", "email"],
		"gender":    {"one_of": ["male", "female"]},
		"phone":     {"max_length": 11},
		"password":  ["required", {"min_length": 10}],
		"password2": {"equal_to_field": "password"}
	}`

	var rules map[string]interface{}
	err := json.Unmarshal([]byte(jsonRules), &rules)
	if err != nil {
		panic(err)
	}

	var jsonData = []byte(`{
		"name": "Jekyll",
		"email": "dangerous.game@dregs.us",
		"gender": "male",
		"phone": "12025550193",
		"password": "take_me_as_i_am",
		"password2": "take_me_as_i_am"
	}`)

	var data map[string]interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	validator := livr.New(&livr.Options{LivrRules: rules})
	validatedData, err := validator.Validate(data)
	if err != nil {
		panic(validator.Errors())
	}

	fmt.Println(validatedData)
}
```

If you want, you can get rid of json unmarshal.
```go
type d livr.Dictionary

var rules = d{
	"name": d{"nested_object": "required"},
}
validator := livr.New(&livr.Options{LivrRules: rules})
```

You can use modifiers separately or can combine them with validation:
```go
var jsonRules = `{
    "email": ["required", "trim", "email", "to_lc"]
}`
```

Feel free to register your own rules.
```go
v := livr.New(&livr.Options{LivrRules: rules})

a := livr.Alias{
	Name:  "strong_password",
	Rules: livr.Dictionary{"min_length": 6},
	Error: "WEAK_PASSWORD",
}
v.RegisterAliasedRule(a)
```

Wrong rules (a rule that is not registered, wrong arguments of a rule) are reported when the
validator is built. `Validate` builds it on the first call and panics with a `*livr.SchemaError`
then; call `Prepare` to get the error before.
```go
v := livr.New(&livr.Options{LivrRules: rules})
if err := v.Prepare(); err != nil {
	panic(err) // for example: Rule no_such_rule not registered
}
```

### LIVR 2.5
The rules added in [LIVR 2.5](https://github.com/koorchik/LIVR/blob/master/drafts/v2.5.md)
([changes](https://livr-spec.org/changes.html#v2-5), [rules](https://livr-spec.org/validation-rules.html))
are supported. LIVR 2.5 only adds rules to LIVR 2.0, every LIVR 2.0 schema keeps working:

- Types and numbers: `boolean`, `not_one_of`, `max_decimals`
- Strings: `digits`, `starts_with`, `ends_with`, `contains`
- Dates and times: `iso_date` with bounds (`{"iso_date": {"max": "today"}}`), `iso_datetime`, `iso_time`
- Presence and meta rules: `required_if`, `when`, `values`, `keys`, `any`
- Lists: `list_length`, `unique`
- Modifiers: `split`
- Formats: `uuid`, `ip`, `cidr`, `hostname`, `base64`, `hex`, `digest`, `mongo_id`, `hex_color`, `mac`, `e164`, `semver`, `jwt`
- Checksums: `luhn`, `credit_card`, `iban`

LIVR 2.5 also fixes some rules: the rules of an alias and of `or` see the fields next to the
validated field (so `equal_to_field` compares with a sibling field there), the rules of `list_of`
items see no fields, `iso_date` rejects an argument that is not an object of options
(`{"iso_date": null}` stays the plain `iso_date`), and `url` rejects a dot after the host
(`https://example.com.`), host labels longer than 63 characters and ports above 65535, while
one-digit ports (`https://example.com:8`) pass.

The bounds `today`, `yesterday` and `tomorrow` use the current date in UTC. Go cannot replace the
system clock in tests, so the validator takes a clock (`time.Now` by default):
```go
v := livr.New(&livr.Options{
	LivrRules: rules,
	Now:       func() time.Time { return time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC) },
})
```

## TESTING
1. Clone and update subomodule with test cases (`test/livr/test_suite` for LIVR 2.0,
`test/livr/test_suite_2.5` for LIVR 2.5)
```sh
git submodule update --init --recursive
```
2. Run tests. The repository has no go.mod, so with Go modules create a temporary one first
(as the CI does; older Go versions need an older fatih/color, see .github/workflows/go.yml).
```sh
go mod init github.com/k33nice/go-livr && go mod tidy
go test ./test -v
```

## DESCRIPTION
See [LIVR Specification](http://livr-spec.org) for detailed documentation and list of supported rules.

### Features:

- [x] Rules are declarative and language independent
- [x] Any number of rules for each field
- [x] Return together errors for all fields
- [x] Excludes all fields that do not have validation rules described
- [x] Has possibility to validate complex hierarchical structures
- [x] Easy to describe and understand rules
- [x] Returns understandable error codes(not error messages)
- [x] Easy to add own rules
- [x] Rules are be able to change results output ("trim", "nested\_object", for example)
- [x] Multipurpose (user input validation, configs validation, contracts programming etc)

## LICENSE
Distributed under MIT License, please see [license](https://github.com/k33nice/go-livr/blob/master/LICENSE) file within the code for more details.


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fk33nice%2Fgo-livr.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fk33nice%2Fgo-livr?ref=badge_large)