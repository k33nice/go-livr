package livr

import (
	"errors"
	"strconv"
	"strings"
)

// listLength - check that validated value is a list with the specified number of items:
// {"list_length": 2}, {"list_length": [1, 10]} or {"list_length": {"min": 1}}.
func listLength(args ...interface{}) Validation {
	args, _ = ruleArgs(args)
	min, max := listLengthArgs(args)

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		list, ok := value.([]interface{})
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		if len(list) < min {
			return nil, errors.New("TOO_FEW_ITEMS")
		}
		if max >= 0 && len(list) > max {
			return nil, errors.New("TOO_MANY_ITEMS")
		}

		return value, nil
	}
}

// listLengthArgs - the minimal and the maximal length of list_length, -1 for no maximum.
func listLengthArgs(args []interface{}) (int, int) {
	const usage = "Rule list_length takes a length, a minimal and a maximal length, or an object with min and max"
	if len(args) == 0 {
		schemaErrorf(usage)
	}

	min, max := 0, -1
	if options, ok := args[0].(Dictionary); ok {
		if len(options) == 0 {
			schemaErrorf(usage)
		}
		for key, value := range options {
			length, ok := countArg(value)
			switch {
			case !ok || (key != "min" && key != "max"):
				schemaErrorf(usage)
			case key == "min":
				min = length
			default:
				max = length
			}
		}
	} else {
		var ok bool
		if min, ok = countArg(args[0]); !ok {
			schemaErrorf(usage)
		}
		max = min
		if len(args) > 1 {
			if max, ok = countArg(args[1]); !ok {
				schemaErrorf(usage)
			}
		}
	}

	if max >= 0 && min > max {
		schemaErrorf("Rule list_length got min greater than max")
	}
	return min, max
}

// unique - check that validated value is a list without duplicates. Without arguments the items
// are primitive values, with keys they are objects that differ in the values of these keys.
func unique(args ...interface{}) Validation {
	args, _ = ruleArgs(args)
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		key, ok := stringArg(arg)
		if !ok {
			schemaErrorf("Rule unique takes key names")
		}
		keys = append(keys, key)
	}

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		list, ok := value.([]interface{})
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}

		// Items are compared by the string forms, so 1 and "1" are duplicates. A wrong item gives
		// FORMAT_ERROR wherever it is in the list.
		seen := make(map[string]bool, len(list))
		hasDuplicates := false
		for _, item := range list {
			id, ok := uniqueID(item, keys)
			if !ok {
				return nil, errors.New("FORMAT_ERROR")
			}
			if id == "" {
				continue
			}
			if seen[id] {
				hasDuplicates = true
			}
			seen[id] = true
		}

		if hasDuplicates {
			return nil, errors.New("NOT_UNIQUE_ITEMS")
		}
		return value, nil
	}
}

// uniqueID - the string that identifies a list item for unique, "" for an item that is not
// compared (an empty item, an empty key value); false for an item of a wrong type.
func uniqueID(item interface{}, keys []string) (string, bool) {
	if isEmpty(item) {
		return "", true
	}

	if len(keys) == 0 {
		s, ok := stringForm(item)
		// "\x00" marks a non-empty id.
		return "\x00" + s, ok
	}

	object, ok := item.(Dictionary)
	if !ok {
		return "", false
	}

	// The length of every key value before it, so that ["a,b"] and ["a", "b"] differ.
	var id strings.Builder
	hasEmptyKeyValue := false
	for _, key := range keys {
		keyValue := object[key]
		if isEmpty(keyValue) {
			hasEmptyKeyValue = true
			continue
		}
		s, ok := stringForm(keyValue)
		if !ok {
			return "", false
		}
		id.WriteString(strconv.Itoa(len(s)))
		id.WriteByte(':')
		id.WriteString(s)
	}

	if hasEmptyKeyValue {
		return "", true
	}
	return id.String(), true
}
