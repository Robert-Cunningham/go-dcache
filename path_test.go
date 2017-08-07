package dcache2

import (
	"fmt"
	"testing"
)

func TestSimplifyPath(t *testing.T) {
	assertSimplifiesTo(t, Path{"c/d/e"}, Path{"/a/b"}, Path{"/a/b/c/d/e"})
	assertSimplifiesTo(t, Path{"c/d/./e"}, Path{"/a/b"}, Path{"/a/b/c/d/e"})
	assertSimplifiesTo(t, Path{"c/d/../e"}, Path{"/a/b"}, Path{"/a/b/c/e"})
	assertSimplifiesTo(t, Path{"c/d/../././e"}, Path{"/a/b"}, Path{"/a/b/c/e"})
	assertSimplifiesTo(t, Path{"c/d/../././../e"}, Path{"/a/b"}, Path{"/a/b/e"})
	assertSimplifiesTo(t, Path{"c/d/.././d/./../e"}, Path{"/a/b"}, Path{"/a/b/c/e"})
}

func assertSimplifiesTo(t *testing.T, relative, pwd, expectedAbsolute Path) {
	if relative.simplifyPath(pwd) != expectedAbsolute {
		t.Fail()
		fmt.Printf("Got %v but expected %v.", relative.simplifyPath(pwd), expectedAbsolute)
	}
}
