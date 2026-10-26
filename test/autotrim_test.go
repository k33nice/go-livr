package test

import (
	"reflect"
	"testing"

	"github.com/k33nice/go-livr"
)

func TestAutoTrim(t *testing.T) {
	rules := livr.Dictionary{
		"code":     "required",
		"password": []interface{}{"required", livr.Dictionary{"min_length": 3.0}},
		"list":     livr.Dictionary{"list_of": []interface{}{"string"}},
		"tags":     "not_empty",
		"address": livr.Dictionary{
			"nested_object": livr.Dictionary{
				"street": livr.Dictionary{"min_length": 5.0},
			},
		},
	}

	v := livr.New(&livr.Options{LivrRules: rules, AutoTrim: livr.DoTrim})
	_, err := v.Validate(livr.Dictionary{
		"code":     "  ",
		"password": " 12  ",
		"address":  livr.Dictionary{"street": "  hell "},
	})
	if err == nil {
		t.Fatal("Validation pass but must fail")
	}
	expectedErrs := map[string]interface{}{
		"code":     "REQUIRED",
		"password": "TOO_SHORT",
		"address":  map[string]interface{}{"street": "TOO_SHORT"},
	}
	if errs := indirectErrors(v.Errors()); !reflect.DeepEqual(errs, expectedErrs) {
		t.Errorf("got errors %v, want %v", errs, expectedErrs)
	}

	v = livr.New(&livr.Options{LivrRules: rules, AutoTrim: livr.DoTrim})
	output, err := v.Validate(livr.Dictionary{
		"code":     " A ",
		"password": " 123  ",
		"list":     []interface{}{" aaa ", " bbb "},
		"tags":     []interface{}{},
		"address":  livr.Dictionary{"street": "  hello "},
	})
	if err != nil {
		t.Fatal(v.Errors())
	}
	expected := livr.Dictionary{
		"code":     "A",
		"password": "123",
		"list":     []interface{}{"aaa", "bbb"},
		"tags":     []interface{}{},
		"address":  livr.Dictionary{"street": "hello"},
	}
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("got output %v, want %v", output, expected)
	}
}
