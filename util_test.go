package dcache2

import (
	"testing"
	"fmt"
)

func TestILog2(t *testing.T) {

	if Ilog2(8) != 3 {
		t.Error(fmt.Sprintf("Ilog2(8) is %v", Ilog2(8)))
	}
	if Roundup_pow_of_two(9) != 4 {
		t.Error(fmt.Sprintf("Roundup_pow_of_two(9) is %v", Roundup_pow_of_two(9)))
	}
	if Roundup_pow_of_two(7) != 3 {
		t.Error(fmt.Sprintf("Roundup_pow_of_two(7) is %v", Roundup_pow_of_two(9)))
	}
}