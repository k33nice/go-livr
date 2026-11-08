package livr

import (
	"errors"
	"regexp"
	"strings"
)

// formatValidation - check the string form of validated value with isValid. The output is the
// string form. A boolean gives the error code, an object or a list gives FORMAT_ERROR.
func formatValidation(code string, isValid func(string) bool) Validation {
	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		if _, ok := value.(bool); ok {
			return nil, errors.New(code)
		}
		s, ok := stringForm(value)
		if !ok {
			return nil, errors.New("FORMAT_ERROR")
		}
		if !isValid(s) {
			return nil, errors.New(code)
		}

		return s, nil
	}
}

// formatRule - a format rule without arguments.
func formatRule(name, code string, isValid func(string) bool) Builder {
	return func(args ...interface{}) Validation {
		noArgs(name, args)
		return formatValidation(code, isValid)
	}
}

// versionArg - the argument of ip and cidr: 0 without arguments, 4 for "v4", 6 for "v6".
func versionArg(name string, args []interface{}) int {
	args, _ = ruleArgs(args)
	if len(args) == 0 {
		return 0
	}
	switch args[0] {
	case "v4":
		return 4
	case "v6":
		return 6
	}
	schemaErrorf(`Rule %s takes no arguments, "v4" or "v6"`, name)
	return 0
}

var (
	uuidRe      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-([1-8])[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	nilOrMaxRe  = regexp.MustCompile(`^(?:00000000-0000-0000-0000-000000000000|[fF]{8}-[fF]{4}-[fF]{4}-[fF]{4}-[fF]{12})$`)
	ipv4Re      = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$`)
	hexGroupRe  = regexp.MustCompile(`^[0-9a-fA-F]{1,4}$`)
	prefixRe    = regexp.MustCompile(`^(?:0|[1-9][0-9]{0,2})$`)
	labelsRe    = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	base64Re    = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)
	base64URLRe = regexp.MustCompile(`^(?:[A-Za-z0-9_-]{4})*(?:[A-Za-z0-9_-]{2}(?:==)?|[A-Za-z0-9_-]{3}=?)?$`)
	hexRe       = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	mongoIDRe   = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)
	hexColorRe  = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3,4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	mac48Re     = regexp.MustCompile(`^[0-9a-fA-F]{2}(?:(?::[0-9a-fA-F]{2}){5}|(?:-[0-9a-fA-F]{2}){5})$`)
	mac64Re     = regexp.MustCompile(`^[0-9a-fA-F]{2}(?:(?::[0-9a-fA-F]{2}){7}|(?:-[0-9a-fA-F]{2}){7})$`)
	e164Re      = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
	semverRe    = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	jwtRe       = regexp.MustCompile(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]*$`)
)

// uuid - check that validated value is a UUID (RFC 9562): {"uuid": "v4"} or {"uuid": ["v4", "v7"]}
// for versions; without versions any version, the nil and the max UUID.
func uuid(args ...interface{}) Validation {
	versions, _ := ruleArgs(args)
	allowed := ""
	for _, version := range versions {
		v, ok := version.(string)
		if !ok || len(v) != 2 || v[0] != 'v' || v[1] < '1' || v[1] > '8' {
			schemaErrorf(`Rule uuid takes versions "v1"-"v8"`)
		}
		allowed += v[1:]
	}

	return formatValidation("WRONG_UUID", func(s string) bool {
		m := uuidRe.FindStringSubmatch(s)
		if allowed == "" {
			return m != nil || nilOrMaxRe.MatchString(s)
		}
		return m != nil && strings.Contains(allowed, m[1])
	})
}

// ip - check that validated value is an IPv4 or IPv6 address: {"ip": "v4"}, {"ip": "v6"}.
func ip(args ...interface{}) Validation {
	version := versionArg("ip", args)

	return formatValidation("WRONG_IP", func(s string) bool {
		return (version != 6 && ipv4Re.MatchString(s)) || (version != 4 && isIPv6(s))
	})
}

// cidr - check that validated value is an IP address with a prefix length: "10.0.0.0/8".
func cidr(args ...interface{}) Validation {
	version := versionArg("cidr", args)

	return formatValidation("WRONG_CIDR", func(s string) bool {
		slash := strings.IndexByte(s, '/')
		if slash < 0 || !prefixRe.MatchString(s[slash+1:]) {
			return false
		}
		address, prefix := s[:slash], s[slash+1:]

		maxPrefix := ""
		if version != 6 && ipv4Re.MatchString(address) {
			maxPrefix = "32"
		} else if version != 4 && isIPv6(address) {
			maxPrefix = "128"
		} else {
			return false
		}
		// Decimals without leading zeros compare as their lengths, then as strings.
		return len(prefix) < len(maxPrefix) || (len(prefix) == len(maxPrefix) && prefix <= maxPrefix)
	})
}

// isIPv6 - the text forms of RFC 4291, section 2.2: eight groups, or fewer with one "::" for one
// or more groups of zeros; the last two groups may be an IPv4 address.
func isIPv6(s string) bool {
	gap := strings.Index(s, "::")
	if gap < 0 {
		return countIPv6Groups(s, true) == 8
	}
	// A second "::" (":::" too) is ambiguous.
	if strings.Contains(s[gap+1:], "::") {
		return false
	}
	before := countIPv6Groups(s[:gap], false)
	after := countIPv6Groups(s[gap+2:], true)
	return before >= 0 && after >= 0 && before+after <= 7
}

// countIPv6Groups - the number of 16-bit groups separated by ":", -1 for a wrong group. The last
// one may be an IPv4 address (two groups) at the end of the address.
func countIPv6Groups(s string, isAddressEnd bool) int {
	if s == "" {
		return 0
	}
	groups := strings.Split(s, ":")
	last := len(groups) - 1
	for i, group := range groups {
		if hexGroupRe.MatchString(group) {
			continue
		}
		if i == last && isAddressEnd && ipv4Re.MatchString(group) {
			return len(groups) + 1
		}
		return -1
	}
	return len(groups)
}

// hostname - check that validated value is a hostname (RFC 1123).
var hostname = formatRule("hostname", "WRONG_HOSTNAME", func(s string) bool {
	if len(s) > 253 || !labelsRe.MatchString(s) {
		return false
	}
	// The last label is not all digits, so an IPv4 address is not a hostname.
	for i := len(s) - 1; i >= 0 && s[i] != '.'; i-- {
		if !isDigit(s[i]) {
			return true
		}
	}
	return false
})

// base64 - check that validated value is Base64 (RFC 4648), {"base64": "url"} for the URL-safe
// alphabet with optional padding.
func base64(args ...interface{}) Validation {
	args, _ = ruleArgs(args)
	re := base64Re
	if len(args) > 0 {
		if args[0] != "url" {
			schemaErrorf(`Rule base64 takes no arguments or "url"`)
		}
		re = base64URLRe
	}

	return formatValidation("WRONG_BASE64", re.MatchString)
}

// hex - check that validated value is hex digits.
var hex = formatRule("hex", "WRONG_HEX", hexRe.MatchString)

var digestLengths = map[string]int{"md5": 32, "sha1": 40, "sha256": 64, "sha384": 96, "sha512": 128}

// digest - check that validated value is a hex digest of one of the specified algorithms:
// {"digest": "sha256"}, {"digest": ["md5", "sha1"]}.
func digest(args ...interface{}) Validation {
	algorithms, _ := ruleArgs(args)
	if len(algorithms) == 0 {
		schemaErrorf("Rule digest takes one or more algorithms")
	}

	lengths := make(map[int]bool)
	for _, algorithm := range algorithms {
		name, _ := algorithm.(string)
		length, ok := digestLengths[name]
		if !ok {
			schemaErrorf("Rule digest takes md5, sha1, sha256, sha384 or sha512")
		}
		lengths[length] = true
	}

	return formatValidation("WRONG_DIGEST", func(s string) bool {
		return lengths[len(s)] && hexRe.MatchString(s)
	})
}

// mongoID - check that validated value is a BSON ObjectId.
var mongoID = formatRule("mongo_id", "WRONG_MONGO_ID", mongoIDRe.MatchString)

// hexColor - check that validated value is a hex color: "#fff", "#ffffff80".
var hexColor = formatRule("hex_color", "WRONG_HEX_COLOR", hexColorRe.MatchString)

// mac - check that validated value is a MAC-48 address, {"mac": 64} for EUI-64.
func mac(args ...interface{}) Validation {
	args, _ = ruleArgs(args)
	re := mac48Re
	if len(args) > 0 {
		switch bits, _ := stringArg(args[0]); bits {
		case "48":
		case "64":
			re = mac64Re
		default:
			schemaErrorf("Rule mac takes no arguments, 48 or 64")
		}
	}

	return formatValidation("WRONG_MAC", re.MatchString)
}

// e164 - check that validated value is a phone number in the E.164 format: "+380441234567".
var e164 = formatRule("e164", "WRONG_E164", e164Re.MatchString)

// semver - check that validated value is a version of Semantic Versioning 2.0.0.
var semver = formatRule("semver", "WRONG_SEMVER", semverRe.MatchString)

// jwt - check the structure of a JWT (JWS compact serialization), the segments are not decoded.
var jwt = formatRule("jwt", "WRONG_JWT", func(s string) bool {
	if !jwtRe.MatchString(s) {
		return false
	}
	// No base64url segment has a length of 1 modulo 4.
	for _, segment := range strings.Split(s, ".") {
		if len(segment)%4 == 1 {
			return false
		}
	}
	return true
})
