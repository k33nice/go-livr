package livr

import "regexp"

// luhn - check that validated value is at least 2 digits with the Luhn checksum.
var luhn = formatRule("luhn", "WRONG_CHECKSUM", func(s string) bool {
	return len(s) >= 2 && isLuhn(s)
})

// creditCard - check that validated value is a card number: 12-19 digits with the Luhn checksum.
var creditCard = formatRule("credit_card", "WRONG_CREDIT_CARD_NUMBER", func(s string) bool {
	return len(s) >= 12 && len(s) <= 19 && isLuhn(s)
})

// isLuhn - digits with the Luhn mod 10 checksum 0.
func isLuhn(s string) bool {
	sum := 0
	for i := 0; i < len(s); i++ {
		c := s[len(s)-1-i]
		if !isDigit(c) {
			return false
		}
		d := int(c - '0')
		if i%2 == 1 {
			if d *= 2; d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

var ibanRe = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$`)

// iban - check that validated value is an IBAN (ISO 13616) with the mod 97 checksum.
var iban = formatRule("iban", "WRONG_IBAN", func(s string) bool {
	if !ibanRe.MatchString(s) {
		return false
	}
	// The first four characters move to the end, a letter is two digits: A = 10 ... Z = 35.
	rest := 0
	for _, c := range s[4:] + s[:4] {
		if c >= 'A' {
			rest = (rest*100 + int(c-'A') + 10) % 97
		} else {
			rest = (rest*10 + int(c-'0')) % 97
		}
	}
	return rest == 1
})
