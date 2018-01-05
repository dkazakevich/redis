package testutil

import (
	"testing"
	"reflect"
)

func AssertEquals(t *testing.T, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Fatalf("Expected: '%v'. Actual: '%v'", expected, actual)
	}
}

func AssertEqualsDeep(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected: '%v'. Actual: '%v'", expected, actual)
	}
}
