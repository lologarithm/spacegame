package main

import (
	"fmt"
	"testing"
)

func TestVect(t *testing.T) {
	v1 := &Vect2{2, 2}
	v_mult := MultVect2(v1, 2)
	if v_mult.X() != 4 && v_mult.Y() != 4 {
		t.FailNow()
	}

	fmt.Println("Vector tests pass.")
}
