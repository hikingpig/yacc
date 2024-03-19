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
	p.first = p.prod[p.off] // ================== SHOULD EXPLICITLY INCREASE?

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

// the production stored in pp is the same or missing the first symbol (LHS)
// compared to the original production in prdptr
// the end symbol of a production is negative. it represents an action.
// whenever it encounters negative, it knows to reduce and take action!
func writem(pp Pitem) string {
	p0 := prdptr[pp.prodno] // original production
	p := pp.prod
	q := nonterms[p0[0]-NTBASE].name + ": " // the LHS of p0
	i := len(p) - len(p0) + 1               // idx starts at 0: missing LHS, 1: equal (skip LHS)
	for i < len(p)-1 {                      // skip the end symbol(action)
		c := ' '
		if i == pp.off { // write a '.' at p.off
			c = '.'
		}
		q += string(c)

		q += symName(p[i]) // ================= is it necessary? symname are unquoted!!!
		i++
	}

	n := p[pp.off]
	if n < 0 { // end of production, print action, gonna reduce soon!
		q += fmt.Sprintf("    (%v)", -n)
	}
	return q
}
