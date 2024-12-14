// Package test provides an interface for writing simple tests with ease.
package test

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// New creates a new [Test] to exercise a function f.
// If f is not a function, attempts to evaluate the test will panic.
func New(t *testing.T, f any) *Test {
	return &Test{t: t, f: reflect.ValueOf(f)}
}

// Test describes a set of test cases for a given function.
type Test struct {
	t      *testing.T
	f      reflect.Value
	cases  []testCase
	pcases []testCase
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
	if len(t.pcases) > 0 {
		// if there is an incomplete test case, augment it with input arguments
		if c := t.pcases[len(t.pcases)-1]; c.input == nil {
			c.input = input
			t.cases = append(t.cases, c)
			t.pcases = t.pcases[:len(t.pcases)-1]
			return t
		}
	}
	t.pcases = append(t.pcases, testCase{input: input})
	return t
}

// Want specifies output results that the function is expected to return.
// If there was an [On] call, attaches to it to produce a test case.
func (t *Test) Want(results ...any) *Test {
	want := makeValues(results)
	if len(t.pcases) > 0 {
		// if there is an incomplete test case, augment it with expected results
		if c := t.pcases[len(t.pcases)-1]; c.want == nil {
			c.want = want
			t.cases = append(t.cases, c)
			t.pcases = t.pcases[:len(t.pcases)-1]
			return t
		}
	}
	t.pcases = append(t.pcases, testCase{want: want})
	return t
}

func makeValues(v []any) []reflect.Value {
	args := make([]reflect.Value, 0, len(v))
	for _, val := range v {
		args = append(args, reflect.ValueOf(val))
	}
	return args
}

// AssertEqual checks whether the function passes the [Test] by comparing expected and actual results for equality.
func (t *Test) AssertEqual() {
	if len(t.pcases) > 0 {
		for _, c := range t.pcases {
			t.t.Logf("malformed test case: %v(%v), want {%v}", funcName(t.f), c.input, c.want)
		}
		t.t.FailNow()
	}

	for _, test := range t.cases {
		got := testValues(t.f.Call(test.input))
		dumpFuncCall := func() {
			t.t.Errorf("%v(%v) = {%v}, want {%v}", funcName(t.f), test.input, got, test.want)
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
