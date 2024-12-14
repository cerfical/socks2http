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

	// IsNil checks for nil values.
	IsNil Checker = isNilCheck{}
)

type notNilCheck struct{}

func (notNilCheck) Check(v any) bool {
	return v != nil
}

func (notNilCheck) String() string {
	return "<not-nil>"
}

type isNilCheck struct{}

func (isNilCheck) Check(v any) bool {
	return v == nil
}

func (isNilCheck) String() string {
	return "<nil>"
}
