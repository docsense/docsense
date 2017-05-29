package stemmer

import (
	"testing"
	"fmt"
	"strings"
)

func TestRemovePolish(t *testing.T) {
	s := "ZAŻÓŁĆ GĘŚLĄ JAŹŃ zażółć gęślą jaźń"
	r := "zazolc gesla jazn zazolc gesla jazn"
	x := removepolish(strings.ToLower(s))
	fmt.Println(r, x)
	if x != r {
		t.FailNow()
	}
}
