package test

import (
	"encoding/json"
	"testing"

	"github.com/k33nice/go-livr"
)

// validateJSON - return the output or the errors of the validation as JSON.
func validateJSON(rules, input livr.Dictionary) string {
	v := livr.New(&livr.Options{LivrRules: rules})
	var res interface{}
	if output, err := v.Validate(input); err == nil {
		res = output
	} else {
		res = map[string]interface{}{"errors": indirectErrors(v.Errors())}
	}
	data, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func TestVariableObject(t *testing.T) {
	sel := []interface{}{"type", livr.Dictionary{
		"1":    livr.Dictionary{"type": "required"},
		"0":    livr.Dictionary{"type": "required"},
		"true": livr.Dictionary{"type": "required"},
		"":     livr.Dictionary{"type": "required"},
	}}
	rules := livr.Dictionary{
		"v1": livr.Dictionary{"variable_object": sel},
		"v2": livr.Dictionary{"variable_object": sel},
		"v3": livr.Dictionary{"variable_object": sel},
		"v4": livr.Dictionary{"variable_object": sel},
		"v5": livr.Dictionary{"variable_object": sel},
		"v6": livr.Dictionary{"variable_object": sel},
		"l1": livr.Dictionary{"list_of_different_objects": sel},
		"l2": livr.Dictionary{"list_of_different_objects": sel},
		"l3": livr.Dictionary{"list_of_different_objects": sel},
		"l4": livr.Dictionary{"list_of_different_objects": sel},
	}

	// Empty and missing values are skipped, a number selects the rules by its string form.
	got := validateJSON(rules, livr.Dictionary{
		"v1": nil,
		"v2": "",
		"v3": livr.Dictionary{"type": 1.0},
		"v4": livr.Dictionary{"type": true},
		"l1": nil,
		"l2": "",
		"l3": []interface{}{livr.Dictionary{"type": 1.0}, livr.Dictionary{"type": true}},
	})
	want := `{"l1":null,"l2":"","l3":[{"type":1},{"type":true}],"v1":null,"v2":"","v3":{"type":1},"v4":{"type":true}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	// Empty, zero and false selector values are not looked up, a list of objects
	// is not an object and a single object is not a list.
	got = validateJSON(rules, livr.Dictionary{
		"v1": livr.Dictionary{"type": 0.0},
		"v2": livr.Dictionary{"type": ""},
		"v3": livr.Dictionary{"type": false},
		"v4": livr.Dictionary{"type": 2.0},
		"v5": []interface{}{livr.Dictionary{"type": 1.0}},
		"v6": livr.Dictionary{},
		"l1": livr.Dictionary{"type": 1.0},
		"l2": "abc",
		"l3": []interface{}{livr.Dictionary{"type": 0.0}, livr.Dictionary{"type": 1.0}},
		"l4": []interface{}{livr.Dictionary{"type": false}, "abc"},
	})
	want = `{"errors":{"l1":"FORMAT_ERROR","l2":"FORMAT_ERROR","l3":["FORMAT_ERROR",null],` +
		`"l4":["FORMAT_ERROR","FORMAT_ERROR"],"v1":"FORMAT_ERROR","v2":"FORMAT_ERROR",` +
		`"v3":"FORMAT_ERROR","v4":"FORMAT_ERROR","v5":"FORMAT_ERROR","v6":"FORMAT_ERROR"}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
