package imf

import "strings"

// RFC 5322 3.2.3. Atom.

func IsAtom(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if !isAtomChar(s[i]) {
			return false
		}
	}

	return true
}

func IsDotAtom(s string) bool {
	for _, part := range strings.Split(s, ".") {
		if !IsAtom(part) {
			return false
		}
	}

	return true
}

func isAtomChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '\'' ||
		c == '*' || c == '+' || c == '-' || c == '/' || c == '=' || c == '?' ||
		c == '^' || c == '_' || c == '`' || c == '{' || c == '|' || c == '}' ||
		c == '~'
}
