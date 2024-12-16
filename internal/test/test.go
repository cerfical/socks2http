// Package test provides an interface for writing simple tests with ease.
package test

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/internal/test/checks"
)

// Test describes a set of test cases for a function.
type Test struct {
	cases []testCase
}

type testCase struct {
	input []any
	want  []any
}

func strSlice[T any](vs []T) string {
	builder := strings.Builder{}
	for i, v := range vs {
		if i != 0 {
			fmt.Fprint(&builder, ", ")
		}
		fmt.Fprint(&builder, strAny(v))
	}
	return builder.String()
}

func strAny(v any) string {
	switch v := v.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case error:
		return fmt.Sprintf("errors.New(%q)", v)
	case reflect.Value:
		return strAny(valueToAny(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// On specifies input arguments to pass to the function.
// If there was a [Want] call, attaches to it to produce a test case.
func (t *Test) On(args ...any) *Test {
	if len(t.cases) > 0 {
		// if there is an incomplete test case, augment it with input arguments
		if c := &t.cases[len(t.cases)-1]; c.input == nil {
			c.input = args
			return t
		}
	}
	t.cases = append(t.cases, testCase{input: args})
	return t
}

// Want specifies output results that the function is expected to return.
// If there was an [On] call, attaches to it to produce a test case.
func (t *Test) Want(results ...any) *Test {
	if len(t.cases) > 0 {
		// if there is an incomplete test case, augment it with expected results
		if c := &t.cases[len(t.cases)-1]; c.want == nil {
			c.want = results
			return t
		}
	}
	t.cases = append(t.cases, testCase{want: results})
	return t
}

func makeValues(v []any) []reflect.Value {
	args := make([]reflect.Value, 0, len(v))
	for _, val := range v {
		args = append(args, reflect.ValueOf(val))
	}
	return args
}

// Case adds a simple test case with a single input and a single result.
func (t *Test) Case(arg, result any) *Test {
	return t.On(arg).Want(result)
}

// Assert runs test cases for a function without exiting on failure.
// If f is not a function, Assert panics.
func (test *Test) Assert(t *testing.T, f any) {
	fv := reflect.ValueOf(f)
	for _, test := range test.cases {
		results := fv.Call(makeValues(test.input))
		t.Logf("%v(%v)\ngot: %v\nwant: %v", funcName(fv), strSlice(test.input), strSlice(results), strSlice(test.want))

		if len(test.want) != len(results) {
			t.Fail()
			continue
		}

		for i, got := range results {
			want := test.want[i]
			if want == nil {
				if got.IsNil() {
					continue
				}
				t.Fail()
				break
			}

			if c, ok := want.(checks.Checker); ok {
				if c.Check(valueToAny(got)) {
					continue
				}
				t.Fail()
				break
			}

			if want := reflect.ValueOf(want).Convert(got.Type()); !got.Equal(want) {
				t.Fail()
				break
			}
		}
	}
}

func valueToAny(v reflect.Value) any {
	if v.IsValid() {
		return v.Interface()
	}
	return nil
}

func funcName(f reflect.Value) string {
	rf := runtime.FuncForPC(f.Pointer())
	if rf == nil {
		return "UnknownFunction"
	}
	fullName := rf.Name()

	// only keep function name, remove module prefix
	i := strings.LastIndex(fullName, ".")
	return fullName[i+1:]
}
