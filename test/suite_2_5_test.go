package test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/k33nice/go-livr"
)

// The LIVR 2.5 test suite.
const suite25 = "livr/test_suite_2.5"

type testCase struct {
	rules   livr.Dictionary
	input   livr.Dictionary
	output  interface{}
	errors  interface{}
	aliases []livr.Alias
	now     func() time.Time
}

func readCase(t *testing.T, dir string) testCase {
	var c testCase
	var options struct {
		Now string `json:"now"`
	}
	for name, target := range map[string]interface{}{
		"rules": &c.rules, "input": &c.input, "output": &c.output, "errors": &c.errors,
		"aliases": &c.aliases, "options": &options,
	} {
		data, err := ioutil.ReadFile(filepath.Join(dir, name+".json"))
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatalf("%s.json: %v", name, err)
		}
	}

	if options.Now != "" {
		now, err := time.Parse(time.RFC3339, options.Now)
		if err != nil {
			t.Fatal(err)
		}
		c.now = func() time.Time { return now }
	}
	return c
}

func (c testCase) validator() *livr.Validator {
	v := livr.New(&livr.Options{LivrRules: c.rules, Now: c.now})
	for _, a := range c.aliases {
		v.RegisterAliasedRule(a)
	}
	return v
}

// eachCase - run test for every case of the suite of the 2.5 test suite.
func eachCase(t *testing.T, suite string, test func(*testing.T, testCase)) {
	dirs, err := ioutil.ReadDir(filepath.Join(suite25, suite))
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		path := filepath.Join(suite25, suite, dir.Name())
		t.Run(dir.Name(), func(t *testing.T) {
			test(t, readCase(t, path))
		})
	}
}

func testPositive25(t *testing.T, c testCase) {
	v := c.validator()
	output, err := v.Validate(c.input)
	if err != nil {
		t.Fatalf("got errors %v", indirectErrors(v.Errors()))
	}
	// The output is compared strictly: a boolean is not "true", a number is not its string.
	if !reflect.DeepEqual(c.output, map[string]interface{}(output)) {
		t.Errorf("got output %s, want %s", toJSON(output), toJSON(c.output))
	}
}

func testNegative25(t *testing.T, c testCase) {
	v := c.validator()
	output, err := v.Validate(c.input)
	if err == nil {
		t.Fatalf("validation passed with output %s, must fail", toJSON(output))
	}
	if errs := indirectErrors(v.Errors()); !reflect.DeepEqual(c.errors, errs) {
		t.Errorf("got errors %s, want %s", toJSON(errs), toJSON(c.errors))
	}
}

func TestPositive25(t *testing.T)        { eachCase(t, "positive", testPositive25) }
func TestNegative25(t *testing.T)        { eachCase(t, "negative", testNegative25) }
func TestAliasesPositive25(t *testing.T) { eachCase(t, "aliases_positive", testPositive25) }
func TestAliasesNegative25(t *testing.T) { eachCase(t, "aliases_negative", testNegative25) }

func TestSchemaErrors25(t *testing.T) {
	eachCase(t, "schema_errors", func(t *testing.T, c testCase) {
		v := c.validator()
		err := v.Prepare()
		if err == nil {
			t.Fatal("the validator was built, must fail")
		}
		if _, ok := err.(*livr.SchemaError); !ok {
			t.Fatalf("got %T, want *livr.SchemaError", err)
		}
		// Validate panics with the same error.
		defer func() {
			if r := recover(); r != err {
				t.Errorf("Validate panicked with %v, want %v", r, err)
			}
		}()
		v.Validate(c.input)
	})
}

func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
