package dcache2

import (
	"testing"
)

func TestLeadingSlashes(t *testing.T){
	if remove_leading_slashes("///d1///d2") != "d1///d2" {
		t.Error()
	}
}

func TestLookup(t *testing.T) {

}

func TestDetectLastType(t *testing.T) {
	if detectLastType("../asdf/asdf") != LAST_DOUBLEDOT {
		t.Error()
	}
	if detectLastType("./asdf/asdf") != LAST_DOT {
		t.Error()
	}
	if detectLastType("asdf/asdf") != LAST_NORMAL {
		t.Error()
	}
}


