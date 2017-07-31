package dcache2

import (
	"testing"
	"fmt"
)

func TestGetLastComponentLengthAndHashFromPath(t *testing.T){
	path := "abcdef/asdf"
	hash, length := getLastComponentLengthAndHashFromPath(path)
	
	if length != 6 {t.Error("Length incorrect")}

	if hash != 3039217368 {t.Error("Hash incorrect " + fmt.Sprintf("%v", hash))}
}