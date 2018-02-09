## Validator LIVR
Validator LIVR - Lightweight validator supporting Language Independent Validation Rules Specification (LIVR).

[![Build Status](https://travis-ci.org/k33nice/go-livr.svg?branch=master)](https://travis-ci.org/k33nice/go-livr)
[![Documentation](https://godoc.org/github.com/k33nice/go-livr?status.svg)](https://godoc.org/github.com/k33nice/go-livr)
[![release](https://img.shields.io/badge/release-v2.0.0-blue.svg)](https://github.com/k33nice/go-livr/releases/tag/v2.0.0)

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

## TESTING
1. Clone and update subomodule with test cases
```sh
git submodule update --init --recursive
```
2. Run tests.
```sh
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
