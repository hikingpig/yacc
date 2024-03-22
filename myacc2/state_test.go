package myacc2

import "testing"

func TestStateGen(t *testing.T) {
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

	stategen()
}
