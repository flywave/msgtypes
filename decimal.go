package msgtypes

import (
	"fmt"
	"math"
)

type Decimal [2]int

func NewDecimal(exponent, value int) Decimal {
	return Decimal{exponent, value}
}

func (n Decimal) slice() []int {
	return n[:]
}

func (n Decimal) Float() float64 {
	return float64(n[1]) * math.Pow10(n[0])
}

func (n Decimal) Int() int {
	if n[0] < 0 {
		return n[1] / pow10(-n[0])
	}
	return n[1] * pow10(n[0])
}

type decimalType struct{}

func (t *decimalType) ConvertExt(v interface{}) interface{} {
	d, ok := v.(*Decimal)
	if !ok {
		panic(fmt.Sprintf("unsupported type %T (%#v)", v, v))
	}
	return d.slice()
}

func (t *decimalType) UpdateExt(dst interface{}, src interface{}) {
	a, ok := src.([]int)
	if !ok {
		panic(fmt.Sprintf("unsupported type %T (%#v)", src, src))
	}

	if len(a) != 2 {
		panic(fmt.Sprintf("invalid decimal size: %v", len(a)))
	}

	copy(dst.(*Decimal)[:], a)
}
