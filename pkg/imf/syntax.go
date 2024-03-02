package imf

import (
	"errors"
	"strings"
)

var ErrInvalidUTF8String = errors.New("invalid utf-8 string")

func IsWSP(c byte) bool {
	// RFC 5234 B.1. Core Rules

	return c == ' ' || c == '\t'
}

func IsSpaceSeparator(c byte) bool {
	return IsWSP(c) || c == '\r'
}

func IsAlphaChar(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}

func IsDigitChar(c byte) bool {
	return c >= '0' && c <= '9'
}

func IsFieldChar(c byte) bool {
	// RFC 5322 3.6.8. Optional Fields

	return (c >= 33 && c <= 57) || (c >= 59 && c <= 126)
}

func IsAtomChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '\'' ||
		c == '*' || c == '+' || c == '-' || c == '/' || c == '=' || c == '?' ||
		c == '^' || c == '_' || c == '`' || c == '{' || c == '|' || c == '}' ||
		c == '~'
}

func IsAtom(s string) bool {
	// RFC 5322 3.2.3. Atom.
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if !IsAtomChar(s[i]) {
			return false
		}
	}

	return true
}

func IsDotAtom(s string) bool {
	// RFC 5322 3.2.3. Atom.
	for _, part := range strings.Split(s, ".") {
		if !IsAtom(part) {
			return false
		}
	}

	return true
}
