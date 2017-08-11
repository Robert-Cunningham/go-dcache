package dcache2

import (
	"fmt"
	"testing"
)

func TestGetTargets(t *testing.T) {
	assertGetTargets(t, CreateOp{Path{"/a/b/c/d"}}, []string{"d"})
	assertGetTargets(t, CreateOp{Path{"/a"}}, []string{"a"})
}

func TestGetTraversals(t *testing.T) {
	assertGetTraversals(t, CreateOp{Path{"/a/b/c/d"}}, []string{"a", "b", "c"})
	assertGetTraversals(t, CreateOp{Path{"/a"}}, []string{})
}

func TestMustWaitOn(t *testing.T) {
	assertMustWaitOn(t, Path{"/a/b/c"}, Path{"/a/b"}, true)
	assertMustWaitOn(t, Path{"/a/b/c"}, Path{"/a/b/d"}, false)
	assertMustWaitOn(t, Path{"/a/b/d"}, Path{"/a"}, true)
	assertMustWaitOn(t, Path{"/a"}, Path{"/b/f/d"}, false)
}

func assertMustWaitOn(t *testing.T, p1 Path, p2 Path, expected bool) {
	c1 := CreateOp{p1}
	c2 := CreateOp{p2}
	if mustWaitOn(c1, c2) != expected {
		fmt.Printf("Failed %v and %v.", p1, p2)
		t.Fail()
	}
}

func assertGetTargets(t *testing.T, c Command, expectedResult []string) {
	out := c.getTargets()
	if len(out) != len(expectedResult) {
		fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
		t.Fail()
		return
	}
	for i := range expectedResult {
		if expectedResult[i] != out[i] {
			fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
			t.Fail()
		}
	}
}

func assertGetTraversals(t *testing.T, c Command, expectedResult []string) {
	out := c.getTraversals()
	if len(out) != len(expectedResult) {
		fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
		t.Fail()
		return
	}
	for i := range expectedResult {
		if expectedResult[i] != out[i] {
			fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
			t.Fail()
		}
	}
}
