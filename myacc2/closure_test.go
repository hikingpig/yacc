package myacc2

import (
	"reflect"
	"testing"
)

func TestClosure(t *testing.T) {
	prods = [][]int{
		[]int{4096, 4097, 1, 0},
		[]int{4097, 4098, -1},
		[]int{4098, 4098, 4099, -2},
		[]int{4098, -3},
		[]int{4099, 4100, -4},
		[]int{4100, 4101, -5},
		[]int{4101, 5, 4, 4101, -6},
		[]int{4101, 4102, -7},
		[]int{4102, 4103, -8},
		[]int{4103, 5, -9},
	}

	yields = [][][]int{
		{prods[0]},
		{prods[1]},
		{prods[2], prods[3]},
		{prods[4]},
		{prods[5]},
		{prods[6], prods[7]},
		{prods[8]},
		{prods[9]},
	}

	firstSets = []lkset{{32}, {32}, {32}, {32}, {32}, {32}, {32}}
	empty = []bool{false, true, true, false, false, false, false, false}

	kernls = []item{
		{1, prods[0], 4097, lkset{0}},
		{2, prods[0], 1, lkset{0}},
		{1, prods[1], -1, lkset{2}},
		{2, prods[2], 4099, lkset{34}},
		{3, prods[2], -2, lkset{34}},
		{2, prods[4], -4, lkset{34}},
		{2, prods[5], -5, lkset{34}},
	}
	kernlp = []int{0, 1, 2, 4, 5, 6, 7}
	wSet = make([]wItem, 57)
	tbitset = 1
	clkset = newLkset()
	tests := []struct {
		n        int
		expected []wItem
	}{
		{0, []wItem{
			{item{
				1,
				[]int{4096, 4097, 1, 0},
				4097,
				lkset{0},
			}, true, false},
			{item{
				1,
				[]int{4097, 4098, -1},
				4098,
				lkset{2},
			}, true, false},
			{item{
				1,
				[]int{4098, 4098, 4099, -2},
				4098,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4098, -3},
				-3,
				lkset{34},
			}, true, false},
		}},
		{1, []wItem{
			{item{
				2,
				[]int{4096, 4097, 1, 0},
				1,
				lkset{0},
			}, true, false},
			{item{
				1,
				[]int{4097, 4098, -1},
				4098,
				lkset{2},
			}, true, false},
			{item{
				1,
				[]int{4098, 4098, 4099, -2},
				4098,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4098, -3},
				-3,
				lkset{34},
			}, true, false},
		}},
		{2, []wItem{
			{item{
				1,
				[]int{4097, 4098, -1},
				-1,
				lkset{2},
			}, true, false},
			{item{
				2,
				[]int{4098, 4098, 4099, -2},
				4099,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4099, 4100, -4},
				4100,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4100, 4101, -5},
				4101,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4101, 5, 4, 4101, -6},
				5,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4101, 4102, -7},
				4102,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4102, 4103, -8},
				4103,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4103, 5, -9},
				5,
				lkset{34},
			}, true, false},
		}},
		{3, []wItem{
			{item{
				3,
				[]int{4098, 4098, 4099, -2},
				-2,
				lkset{34},
			}, true, false},
			{item{
				2,
				[]int{4098, 4098, 4099, -2},
				4099,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4099, 4100, -4},
				4100,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4100, 4101, -5},
				4101,
				lkset{34},
			}, true, false},
			{item{
				1,

				[]int{4101, 5, 4, 4101, -6},
				5,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4101, 4102, -7},
				4102,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4102, 4103, -8},
				4103,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4103, 5, -9},
				5,
				lkset{34},
			}, true, false},
		}},
		{4, []wItem{
			{item{
				2,
				[]int{4099, 4100, -4},
				-4,
				lkset{34},
			}, true, false},
			{item{
				2,
				[]int{4098, 4098, 4099, -2},
				4099,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4099, 4100, -4},
				4100,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4100, 4101, -5},
				4101,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4101, 5, 4, 4101, -6},
				5,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4101, 4102, -7},
				4102,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4102, 4103, -8},
				4103,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4103, 5, -9},
				5,
				lkset{34},
			}, true, false},
		}},
		{5, []wItem{
			{item{
				2,
				[]int{4100, 4101, -5},
				-5,
				lkset{34},
			}, true, false},
			{item{
				2,
				[]int{4098, 4098, 4099, -2},
				4099,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4099, 4100, -4},
				4100,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4100, 4101, -5},
				4101,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4101, 5, 4, 4101, -6},
				5,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4101, 4102, -7},
				4102,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4102, 4103, -8},
				4103,
				lkset{34},
			}, true, false},
			{item{
				1,
				[]int{4103, 5, -9},
				5,
				lkset{34},
			}, true, false},
		}},
	}

	for _, test := range tests {
		closure(test.n)
		result := []wItem{}
		for _, w := range wSet {
			if w.item.prd == nil {
				break
			}
			result = append(result, w)
		}
		if len(result) != len(test.expected) {
			t.Errorf("resulting workset doesn' have correct length: expected: %d, got: %d\n", len(test.expected), len(result))
		}
		for i, w := range result {
			if !reflect.DeepEqual(w, test.expected[i]) {
				t.Errorf("closure [%d] not correct: expected: %+v, got %+v\n", i, test.expected[i], w)
			}
		}
	}

}
