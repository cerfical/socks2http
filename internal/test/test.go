// Package test provides an interface for writing simple tests with ease.
package test

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Test describes a set of test cases for a function.
type Test struct {
	cases []testCase
}

type testCase struct {
	input testValues
	want  testValues
}

type testValues []reflect.Value

func (vs testValues) String() string {
	builder := strings.Builder{}
	for i, v := range vs {
		if i != 0 {
			fmt.Fprint(&builder, ", ")
		}
		fmt.Fprintf(&builder, "%#v", v)
	}
	return builder.String()
}

// On specifies input arguments to pass to the function.
// If there was a [Want] call, attaches to it to produce a test case.
func (t *Test) On(args ...any) *Test {
	input := makeValues(args)
	if len(t.cases) > 0 {
		// if there is an incomplete test case, augment it with input arguments
		if c := &t.cases[len(t.cases)-1]; c.input == nil {
			c.input = input
			return t
		}
	}
	t.cases = append(t.cases, testCase{input: input})
	return t
}

// Want specifies output results that the function is expected to return.
// If there was an [On] call, attaches to it to produce a test case.
func (t *Test) Want(results ...any) *Test {
	want := makeValues(results)
	if len(t.cases) > 0 {
		// if there is an incomplete test case, augment it with expected results
		if c := &t.cases[len(t.cases)-1]; c.want == nil {
			c.want = want
			return t
		}
	}
	t.cases = append(t.cases, testCase{want: want})
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

// AssertEqual checks whether the function satisfies the test by comparing expected and actual results for equality.
// If f is not a function, AssertEqual panics.
func (test *Test) AssertEqual(t *testing.T, f any) {
	fv := reflect.ValueOf(f)
	for _, test := range test.cases {
		got := testValues(fv.Call(test.input))
		dumpFuncCall := func() {
			t.Errorf("%v(%v) = {%v}, want {%v}", funcName(fv), test.input, got, test.want)
		}

		if len(test.want) != len(got) {
			dumpFuncCall()
			continue
		}

		for i := range got {
			if !got[i].Equal(test.want[i]) {
				dumpFuncCall()
				break
			}
		}
	}
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
