package base

import "fmt"

// ToString formats a value without changing the existing container String methods.
// Nil-like values, as defined by IsNil, return "<nil>" without invoking methods.
// Otherwise fmt.Formatter, error and fmt.Stringer take precedence through
// fmt.Sprint. A value implementing only StringerInterface uses its ToString
// method; all other values use fmt.Sprint's default formatting, not %+v.
// Legacy ToString panics propagate; standard formatting retains fmt's behavior.
// Dispatch applies only to the top-level value, not recursively to members.
// An embedded Base's promoted String method takes precedence over an outer
// ToString method. Override String as well when changing that representation.
func ToString(value any) string {
	if IsNil(value) {
		return "<nil>"
	}
	switch value.(type) {
	case fmt.Formatter, error, fmt.Stringer:
		return fmt.Sprint(value)
	}
	if legacy, ok := value.(StringerInterface); ok {
		return legacy.ToString()
	}
	return fmt.Sprint(value)
}
