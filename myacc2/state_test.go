package myacc2

import (
	"reflect"
	"testing"
)

func TestStateGen(t *testing.T) {
	nterm = 6
	nnonterm = 7
	allPrds = []prd{
		{0, []int{4096, 4097, 1, 0}},
		{1, []int{4097, 4098, -1}},
		{2, []int{4098, 4098, 4099, -2}},
		{3, []int{4098, -3}},
		{4, []int{4099, 4100, -4}},
		{5, []int{4100, 4101, -5}},
		{6, []int{4101, 5, 4, 4101, -6}},
		{7, []int{4101, 4102, -7}},
		{8, []int{4102, 4103, -8}},
		{9, []int{4103, 5, -9}},
	}

	prdsStartWith = [][]prd{
		{allPrds[0]},
		{allPrds[1]},
		{allPrds[2], allPrds[3]},
		{allPrds[4]},
		{allPrds[5]},
		{allPrds[6], allPrds[7]},
		{allPrds[8]},
		{allPrds[9]},
	}
	firstSets = []lkset{{32}, {32}, {32}, {32}, {32}, {32}, {32}}
	empty = []bool{false, true, true, false, false, false, false, false}
	wSet = make([]wItem, 57)
	tbitset = 1
	clkset = newLkset()
	expected := []item{
		{1, prd{0, []int{4096, 4097, 1, 0}}, 4097, lkset{0}},
		{2, prd{0, []int{4096, 4097, 1, 0}}, 1, lkset{0}},
		{2, prd{1, []int{4097, 4098, -1}}, -1, lkset{2}},
		{2, prd{2, []int{4098, 4098, 4099, -2}}, 4099, lkset{34}},
		{3, prd{2, []int{4098, 4098, 4099, -2}}, -2, lkset{34}},
		{2, prd{4, []int{4099, 4100, -4}}, -4, lkset{34}},
		{2, prd{5, []int{4100, 4101, -5}}, -5, lkset{34}},
		{2, prd{6, []int{4101, 5, 4, 4101, -6}}, 4, lkset{34}},
		{2, prd{9, []int{4103, 5, -9}}, -9, lkset{34}},
		{2, prd{7, []int{4101, 4102, -7}}, -7, lkset{34}},
		{2, prd{8, []int{4102, 4103, -8}}, -8, lkset{34}},
		{3, prd{6, []int{4101, 5, 4, 4101, -6}}, 4101, lkset{34}},
		{4, prd{6, []int{4101, 5, 4, 4101, -6}}, -6, lkset{34}},
		{2, prd{8, []int{4102, 4103, -8}}, -8, lkset{34}},
		{2, prd{9, []int{4103, 5, -9}}, -9, lkset{34}},
	}
	stategen()
	res := []item{}
	for _, a := range kernls {
		if a.prd.prd != nil {
			res = append(res, a)
		}
	}
	if !reflect.DeepEqual(expected, res) {
		t.Errorf("stategen result conflict: expected: %+v, got: %+v\n", expected, res)
	}
}
