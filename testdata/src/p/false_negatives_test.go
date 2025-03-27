package p

import (
	"reflect"
	"testing"
)

func TestFalseNegatives(t *testing.T) {
	// Test reflection usage
	v := reflect.ValueOf(reflectionFunction)
	result := v.Call(nil)[0].String()
	if result != "reflection" {
		t.Error("unexpected reflection result")
	}

	// Test type assertion usage
	var i interface{} = &TestType{Field: "test"}
	result = i.(*TestType).testMethod()
	if result != "type assertion" {
		t.Error("unexpected type assertion result")
	}
}
