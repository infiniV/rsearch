package parser

import "fmt"

// Position represents a position in the input string
type Position struct {
	Offset int // byte offset
	Line   int // line number (1-indexed)
	Column int // column number (1-indexed)
}

// String returns a string representation of the position
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Message  string
	Position Position
	Line     int
	Column   int
}

// Error implements the error interface
func (e *ParseError) Error() string {
	return fmt.Sprintf("%s at %s", e.Message, e.Position)
}

// NewParseError creates a new parse error
func NewParseError(message string, pos Position) *ParseError {
	return &ParseError{
		Message:  message,
		Position: pos,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

// ParseErrors represents multiple parse errors
type ParseErrors struct {
	Errors []*ParseError
}

// Error implements the error interface
func (e *ParseErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", e.Errors[0].Error(), len(e.Errors)-1)
}

// Add adds a parse error
func (e *ParseErrors) Add(err *ParseError) {
	e.Errors = append(e.Errors, err)
}

// HasErrors returns true if there are any errors
func (e *ParseErrors) HasErrors() bool {
	return len(e.Errors) > 0
}
