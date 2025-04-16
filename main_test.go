// main_test.go
package main

import "testing"

func TestSomething(t *testing.T) {
	expected := "something"
	actual := "something"
	if actual != expected {
		t.Errorf("Expected %s but got %s", expected, actual)
	}
}
