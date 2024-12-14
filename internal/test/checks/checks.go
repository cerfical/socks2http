package checks

import "fmt"

// Checker implements custom rules for validating test results.
type Checker interface {
	fmt.Stringer

	// Check checks a single value returned from a test function call.
	Check(v any) bool
}

var (
	// NotNil checks for non-nil values.
	NotNil Checker = notNilCheck{}

	// Nil checks for nil values.
	Nil Checker = nilCheck{}
)

type notNilCheck struct{}

func (notNilCheck) Check(v any) bool {
	return v != nil
}

func (notNilCheck) String() string {
	return "<not-nil>"
}

type nilCheck struct{}

func (nilCheck) Check(v any) bool {
	return v == nil
}

func (nilCheck) String() string {
	return "<nil>"
}
