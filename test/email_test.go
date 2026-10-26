package test

import (
	"testing"

	"github.com/k33nice/go-livr"
)

func TestEmail(t *testing.T) {
	valid := []string{
		"user@hotline.finance",
		"a#b!c%d/e@example.com",
		`"john doe"@example.com`,
		"user@[192.0.2.1]",
	}
	invalid := []string{
		"a@b@example.com",
		"user@exa_mple.com",
		"a\vb@example.com",
		"\ufeffuser@example.com",
		"a\u00a0b@example.com",
	}

	for _, email := range valid {
		v := livr.New(&livr.Options{LivrRules: livr.Dictionary{"email": "email"}})
		if _, err := v.Validate(livr.Dictionary{"email": email}); err != nil {
			t.Errorf("%q: got errors %v, want none", email, indirectErrors(v.Errors()))
		}
	}
	for _, email := range invalid {
		v := livr.New(&livr.Options{LivrRules: livr.Dictionary{"email": "email"}})
		v.Validate(livr.Dictionary{"email": email})
		if errs := indirectErrors(v.Errors()); !JSONDuckEqual(errs, map[string]interface{}{"email": "WRONG_EMAIL"}) {
			t.Errorf("%q: got errors %v, want WRONG_EMAIL", email, errs)
		}
	}
}
