package myyacc

import "fmt"

type Pitem struct {
	prod   []int
	off    int // offset within the production
	first  int // first term or non-term in item
	prodno int // production number for sorting
}

type Item struct {
	pitem Pitem
	look  Lkset
}

var pidebug = 0     // debugging flag for putitem
var statemem []Item // items for all states

// put an item on statemem
func putitem(p Pitem, set Lkset) {
	p.off++
	p.first = p.prod[p.off]

	if pidebug != 0 && foutput != nil {
		fmt.Fprintf(foutput, "putitem(%v), state %v\n", writem(p), nstate)
	}
	j := pstate[nstate+1]
	if j >= len(statemem) {
		extend(&statemem, STATEINC)
	}
	statemem[j].pitem = p
	if nolook == 0 {
		s := newLkset()
		copy(s, set)
		statemem[j].look = s
	}
	j++
	pstate[nstate+1] = j // ============= next time, for same nstate, j = j+1, will set statemem[j+1]
}

// creates output string for item pointed to by pp
func writem(pp Pitem) string {
	var i int

	p := pp.prod
	q := chcopy(nonterms[prdptr[pp.prodno][0]-NTBASE].name) + ": "
	npi := pp.off

	pi := aryeq(p, prdptr[pp.prodno])

	for {
		c := ' '
		if pi == npi {
			c = '.'
		}
		q += string(c)

		i = p[pi]
		pi++
		if i <= 0 {
			break
		}
		q += chcopy(symName(i))
	}

	// an item calling for a reduction
	i = p[npi]
	if i < 0 {
		q += fmt.Sprintf("    (%v)", -i)
	}

	return q
}
