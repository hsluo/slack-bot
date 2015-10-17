package main

import (
	"fmt"
	"testing"
)

func TestVoteResult(t *testing.T) {
	v1 := newStringSet()
	v1.add("jack")
	v2 := newStringSet()
	v2.add("alice")
	v2.add("bob")
	vr := VoteResult{
		"2": v2,
		"1": v1,
	}
	output := `Result:
1: [jack]
2: [alice bob]
`
	if vr.String() != output {
		fmt.Println(vr)
		t.FailNow()
	}
}
