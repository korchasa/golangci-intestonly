package string_reference

import "testing"

func TestStringRefFunction(t *testing.T) {
	result := StringRefFunction()
	if result != "I appear in strings but am only called in tests" {
		t.Errorf("Unexpected result: %s", result)
	}
}

func TestCommentRefFunction(t *testing.T) {
	result := CommentRefFunction()
	if result != "I'm only used in comments and tests" {
		t.Errorf("Unexpected result: %s", result)
	}
}

func TestReflectionRefFunction(t *testing.T) {
	result := ReflectionRefFunction()
	if result != "I'm used via reflection" {
		t.Errorf("Unexpected result: %s", result)
	}
}
