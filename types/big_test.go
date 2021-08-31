package types

import (
	"testing"
)

func TestMulFloat(t *testing.T) {
	val, val2 := int64(18922020), int64(300)
	a := NewInt(val)

	b := Div(Mul(a, NewInt(2500)), NewInt(10000))
	c := MulFloat(a, 0.25)
	if !b.Equals(c) {
		t.Fatalf("%v not match %v", b, c)
	}
	t.Log(b, c)

	d := NewInt(val2)
	f := DivFloat(a, d)
	g := float64(val) / float64(val2)
	if f != g {
		t.Fatalf("%v not match %v", d, g)
	}
	t.Log(f, g)

	f = DivFloat(d, a)
	g = float64(val2) / float64(val)
	if f != g {
		t.Fatalf("%v not match %v", d, g)
	}
	t.Log(f, g)
}
