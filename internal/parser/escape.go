package parser

import "strings"

// Special characters that need escaping in OpenSearch query syntax
var specialChars = map[rune]bool{
	'+': true, '-': true, '=': true, '&': true, '|': true,
	'>': true, '<': true, '!': true, '(': true, ')': true,
	'{': true, '}': true, '[': true, ']': true, '^': true,
	'"': true, '~': true, '*': true, '?': true, ':': true,
	'\\': true, '/': true, ' ': true,
}

// isSpecialChar returns true if the character needs escaping
func isSpecialChar(ch rune) bool {
	return specialChars[ch]
}

// unescapeString processes escape sequences in a string
// Handles: \+ \- \= \&& \|| \> \< \! \( \) \{ \} \[ \] \^ \" \~ \* \? \: \\ \/
func unescapeString(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			next := rune(s[i+1])
			if isSpecialChar(next) {
				result.WriteRune(next)
				i++ // skip the next character
				continue
			}
		}
		result.WriteByte(s[i])
	}

	return result.String()
}

// escapeString adds escape sequences to special characters
func escapeString(s string) string {
	var result strings.Builder
	result.Grow(len(s) * 2) // worst case: all chars need escaping

	for _, ch := range s {
		if isSpecialChar(ch) {
			result.WriteRune('\\')
		}
		result.WriteRune(ch)
	}

	return result.String()
}

// needsEscaping returns true if the string contains special characters
func needsEscaping(s string) bool {
	for _, ch := range s {
		if isSpecialChar(ch) {
			return true
		}
	}
	return false
}
