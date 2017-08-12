package dcache2

import (
	"testing"
)

type TestOp struct {
	executionFunction func()
	CreateOp
}

func (c TestOp) execute() {
	c.executionFunction()
}

func TestGetNonEmptySubstrings(t *testing.T) {
	out := getNonEmptySubStrings([]string{"a", "b", "c"})
	expected := []string{"a", "ab", "abc"}
	if len(out) != len(expected) {
		t.Fail()
	}
	for i := range out {
		if out[i] != expected[i] {
			t.Fail()
		}
	}
}
