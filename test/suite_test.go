package test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
	"github.com/k33nice/go-livr"
)

var suites interface{}

func files(suite string) []os.FileInfo {
	suites, err := ioutil.ReadDir(filepath.Join("livr/test_suite", suite))
	if err != nil {
		log.Fatal(err)
	}
	return suites
}

func fileBase(p string) string {
	fName := path.Base(p)
	extName := path.Ext(p)
	return fName[:len(fName)-len(extName)]
}

var printS = color.New(color.Bold, color.FgGreen).PrintfFunc()
var fErr = color.New(color.Bold, color.FgRed).SprintfFunc()

func TestPositive(t *testing.T) {
	suites := files("positive")
	for _, suite := range suites {
		files, err := ioutil.ReadDir(filepath.Join("livr/test_suite/positive", suite.Name()))
		if err != nil {
			t.Fatal(err)
		}

		printS("\nRun test for suite %s...\n\n", suite.Name())

		var rules, input, output map[string]interface{}
		for _, f := range files {
			data, err := ioutil.ReadFile(filepath.Join("livr/test_suite/positive", suite.Name(), f.Name()))
			if err != nil {
				t.Fatal(err)
			}

			switch fileBase(f.Name()) {
			case "rules":
				json.Unmarshal(data, &rules)
			case "input":
				json.Unmarshal(data, &input)
			case "output":
				json.Unmarshal(data, &output)
			}

		}
		v := livr.New(&livr.Options{LivrRules: rules})
		r, err := v.Validate(input)
		if err != nil {
			t.Error(v.Errors())
		}
		eq := JSONDuckEqual(output, r)
		if !eq {
			t.Errorf(fErr("FAIL"))
		} else {
			printS("PASS\n")
		}
	}
}

func TestNegative(t *testing.T) {
	suites := files("negative")
	for _, suite := range suites {
		files, err := ioutil.ReadDir(filepath.Join("livr/test_suite/negative", suite.Name()))
		if err != nil {
			t.Fatal(err)
		}

		printS("\n••• Run test for suite %s... •••\n", suite.Name())

		var rules, input, output map[string]interface{}
		for _, f := range files {
			data, err := ioutil.ReadFile(filepath.Join("livr/test_suite/negative", suite.Name(), f.Name()))
			if err != nil {
				t.Fatal(err)
			}

			switch fileBase(f.Name()) {
			case "rules":
				json.Unmarshal(data, &rules)
			case "input":
				json.Unmarshal(data, &input)
			case "errors":
				json.Unmarshal(data, &output)
			}

		}
		v := livr.New(&livr.Options{LivrRules: rules})
		_, err = v.Validate(input)
		if err == nil {
			t.Error("Validation pass but must fail")
		}

		eq := JSONDuckEqual(output, indirectErrors(v.Errors()))
		if !eq {
			t.Errorf(fErr("FAIL"))
		} else {
			printS("PASS\n")
		}
	}
}

func TestAliasesPositive(t *testing.T) {
	suites := files("aliases_positive")
	for _, suite := range suites {
		files, err := ioutil.ReadDir(filepath.Join("livr/test_suite/aliases_positive", suite.Name()))
		if err != nil {
			t.Fatal(err)
		}

		printS("\n••• Run test for suite %s... •••\n", suite.Name())

		var rules, input, output map[string]interface{}
		var aliases []livr.Alias
		for _, f := range files {
			data, err := ioutil.ReadFile(filepath.Join("livr/test_suite/aliases_positive", suite.Name(), f.Name()))
			if err != nil {
				t.Fatal(err)
			}

			switch fileBase(f.Name()) {
			case "rules":
				json.Unmarshal(data, &rules)
			case "input":
				json.Unmarshal(data, &input)
			case "output":
				json.Unmarshal(data, &output)
			case "aliases":
				json.Unmarshal(data, &aliases)
			}

		}
		v := livr.New(&livr.Options{LivrRules: rules})

		for _, a := range aliases {
			v.RegisterAliasedRule(a)
		}
		re, err := v.Validate(input)
		if err != nil {
			t.Error(v.Errors())
		}

		eq := JSONDuckEqual(output, re)
		if !eq {
			t.Error(fErr("FAIL"))
		} else {
			printS("PASS\n")
		}
	}
}

func TestAliasesNegative(t *testing.T) {
	suites := files("aliases_negative")
	for _, suite := range suites {
		files, err := ioutil.ReadDir(filepath.Join("livr/test_suite/aliases_negative", suite.Name()))
		if err != nil {
			t.Fatal(err)
		}

		printS("\n••• Run test for suite %s... •••\n", suite.Name())

		var rules, input, output map[string]interface{}
		var aliases []livr.Alias
		for _, f := range files {
			data, err := ioutil.ReadFile(filepath.Join("livr/test_suite/aliases_negative", suite.Name(), f.Name()))
			if err != nil {
				t.Fatal(err)
			}

			switch fileBase(f.Name()) {
			case "rules":
				json.Unmarshal(data, &rules)
			case "input":
				json.Unmarshal(data, &input)
			case "errors":
				json.Unmarshal(data, &output)
			case "aliases":
				json.Unmarshal(data, &aliases)
			}

		}
		v := livr.New(&livr.Options{LivrRules: rules})

		for _, a := range aliases {
			v.RegisterAliasedRule(a)
		}
		_, err = v.Validate(input)
		if err == nil {
			t.Error("Validation pass but must fail")
		}

		eq := JSONDuckEqual(output, indirectErrors(v.Errors()))
		if !eq {
			t.Errorf(fErr("FAIL"))
		} else {
			printS("PASS\n")
		}
	}
}

func indirectErrors(errs interface{}) interface{} {
	var res interface{}
	switch e := errs.(type) {
	case map[string]interface{}:
		var r = make(map[string]interface{})
		for k, err := range e {
			switch e := err.(type) {
			case map[string]interface{}:
				r[k] = indirectErrors(e)
			case []interface{}:
				s := make([]interface{}, len(e))
				for i, err := range e {
					s[i] = indirectErrors(err)
				}
				r[k] = s
			case interface{}:
				if v, ok := e.(error); ok {
					r[k] = v.Error()
				}
			}
		}
		res = r
	case error:
		res = e.Error()
	}

	return res
}
