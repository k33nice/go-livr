package livr

import (
	"strconv"
	"strings"
)

// parseReference - the segments of a reference to another field: "name", or a path "a/b/0" from
// the containing object. "~1" stands for "/" and "~0" for "~", as in JSON Pointer (RFC 6901).
// A reference that is not a string, is empty, starts or ends with "/", has an empty segment or a
// wrong escape is a schema error.
func parseReference(reference interface{}, ruleName string) []string {
	ref, ok := reference.(string)
	if !ok || ref == "" {
		schemaErrorf("A reference of %s must be a non-empty string", ruleName)
	}

	path := strings.Split(ref, "/")
	for i, segment := range path {
		if segment == "" {
			schemaErrorf("The reference %q of %s has an empty segment", ref, ruleName)
		}
		if !strings.Contains(segment, "~") {
			continue
		}
		for j := 0; j < len(segment); j++ {
			if segment[j] == '~' && (j+1 == len(segment) || (segment[j+1] != '0' && segment[j+1] != '1')) {
				schemaErrorf("The reference %q of %s has a wrong escape", ref, ruleName)
			}
		}
		path[i] = strings.Replace(strings.Replace(segment, "~1", "/", -1), "~0", "~", -1)
	}

	return path
}

// resolveReference - the input value at the path from the containing object, nil when a segment
// is missing or a value on the way is not an object (or a list for an index segment).
func resolveReference(params Dictionary, path []string) interface{} {
	var value interface{} = params
	for _, segment := range path {
		switch v := value.(type) {
		case Dictionary:
			value = v[segment]
		case []interface{}:
			// An index is "0" or a number without leading zeros.
			if segment != "0" && (segment[0] == '0' || segment[0] == '-' || segment[0] == '+') {
				return nil
			}
			i, err := strconv.Atoi(segment)
			if err != nil || i >= len(v) {
				return nil
			}
			value = v[i]
		default:
			return nil
		}
	}
	return value
}
