package dcache2

import (
	"math"
)

func Roundup_pow_of_two(n int) uint {
	return (Ilog2(uint(n - 1)) + 1)
}

func Ilog2(n uint) uint { //Inefficient.
	return uint(math.Log2(float64(n)))
}
const GOLDEN_RATIO_32  = 0x61C88647
const GOLDEN_RATIO_64  = 0x61C8864680B583EB

func hash_32(val uint32, bits int) uint32 {
	return (val * GOLDEN_RATIO_32) >> uint(32 - bits)
}

func convertInterfaceArrayToDentryArray(i []interface{}) ([]*Dentry) {
	out := make([]*Dentry, len(i))
	for a := 0; a < len(i); a++ {
		if (i[a] == nil) {
			continue
		} else {
			out[a] = i[a].(*Dentry)
		}
	}

	return out
}

func removeNilFromSlice(i []interface{}) ([]interface{}) {
	var out []interface{}
	for a := 0; a < len(i); a++ {
		if i[a] != nil {
			out = append(out, i[a])
		}
	}

	return out
}