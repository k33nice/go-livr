package livr

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

var dateReg = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)

// isoDate - make sure that validated value is valid date in format "2006-01-02" ("YYYY-MM-DD").
func isoDate(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		if _, ok := value.(string); !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		t, err := time.Parse("2006-01-02", value.(string))
		if err != nil {
			return nil, errors.New("WRONG_DATE")
		}
		// TODO: it's can be redundant.
		if t.Format("2006-01-02") != value.(string) {
			return nil, errors.New("WRONG_DATE")
		}

		return value, nil
	}
}

var urlRe = regexp.MustCompile(
	`(?i)^(?:(?:http|https)://)(?:\S+(?::\S*)?@)?(?:(?:(?:[1-9]\d?|1\d\d|2[0-1]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z0-9]-*)*[a-z0-9]+)(?:\.(?:[a-z0-9]-*)*[a-z0-9]+)*(?:\.(?:[a-z]{2,})))\.?|localhost)(?::\d{2,5})?(?:[/?#]\S*)?$`,
)

// url - make sure that validated value is valid url.
func url(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		if _, ok := value.(string); !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		if !urlRe.MatchString(value.(string)) {
			return nil, errors.New("WRONG_URL")
		}

		return value, nil
	}
}

var emailRe = regexp.MustCompile(
	`(?i)^([\w\-_+]+(?:\.[\w\-_+]+)*)@((?:[\w\-]+\.)*\w[\w\-]{0,66})\.([a-z]{2,6}(?:\.[a-z]{2})?)$`,
)

// email - make sure that validated value is valid email address.
func email(args ...interface{}) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if value == nil || value == "" {
			return value, nil
		}

		if _, ok := value.(string); !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		if !emailRe.MatchString(value.(string)) {
			return nil, errors.New("WRONG_EMAIL")
		}
		if regexp.MustCompile(`@.*@`).MatchString(value.(string)) {
			return nil, errors.New("WRONG_EMAIL")
		}
		if regexp.MustCompile(`@.*_`).MatchString(value.(string)) {
			return nil, errors.New("WRONG_EMAIL")
		}

		return value, nil
	}
}

// equalToField - make sure that validated value is equal to some filed.
func equalToField(args ...interface{}) Validation {
	var field string
	if len(args) > 0 {
		if v, ok := args[0].(string); ok {
			field = v
		}
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		var params Dictionary
		if len(builders) > 0 {
			if v, ok := builders[0].(Dictionary); ok {
				params = v
			}
		}
		if value == nil || value == "" {
			return value, nil
		}

		var expectedVal interface{}
		if v, ok := params[field]; ok {
			expectedVal = v
		}

		switch v := value.(type) {
		case bool:
			switch vv := expectedVal.(type) {
			case bool:
				if vv == v {
					return value, nil
				}
			case string:
				if vs, _ := strconv.ParseBool(vv); vs == v {
					return value, nil
				}
			}
		case string:
			switch vv := expectedVal.(type) {
			case bool:
				if vb := strconv.FormatBool(vv); vb == v {
					return value, nil
				}
			case string:
				if vv == v {
					return value, nil
				}
			case float64:
				if vf := strconv.FormatFloat(vv, 'f', -1, 64); vf == v {
					return value, nil
				}
			}
		case float64:
			switch vv := expectedVal.(type) {
			case string:
				if vs, _ := strconv.ParseFloat(vv, 64); vs == v {
					return value, nil
				}
			case float64:
				if vv == v {
					return value, nil
				}
			}
		default:
			return nil, errors.New("FORMAT_ERROR")
		}

		return nil, errors.New("FIELDS_NOT_EQUAL")
	}
}
