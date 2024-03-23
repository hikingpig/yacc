package myacc2

import (
	"sort"
)

const STATEINC = 200

var nstate = 0
var stateChain = make([]int, NSTATES) // chain states of the same transition symbol together
var tstates []int                     // latest state generated by a terminal transition symbol
var ntstates []int                    // latest state generated by a non-terminal transition symbol
var nterm int
var nnonterm int

func addKernItem(itemI item) {
	itemI.off++
	itemI.first = itemI.prd.prd[itemI.off]

	i := kernlp[nstate+1]
	if i >= len(kernls) {
		extend(&kernls, STATEINC)
	}
	kernls[i] = itemI.clone()
	kernlp[nstate+1] = i + 1
}

// stategen generates the LALR(1) kernel items of states
func stategen() {
	nstate = 0
	tstates = make([]int, nterm+1)
	ntstates = make([]int, nnonterm+1)
	clkset = newLkset()
	addKernItem(item{0, allPrds[0], 0, newLkset()})
	nstate++
	i := 0
	for i < nstate {
		closure(i)
		for j := 0; j < cwp; j++ {
			w := wSet[j]
			if w.done {
				continue
			}
			w.done = true
			first := w.item.first
			if first <= 1 { // end of prd or $end. skip
				continue
			}
			kernlp[nstate+1] = kernlp[nstate] // next state hold 0 items initially
			addKernItem(w.item)
			for k := j + 1; k < cwp; k++ {
				if first == wSet[k].item.first {
					addKernItem(wSet[k].item)
					wSet[k].done = true
				}
			}
			newState(first)
		}
		i++
	}
}

// newState creates a new state for a transtion symbol or merge it with existing state
func newState(sym int) {
	p1, p2 := kernlp[nstate], kernlp[nstate+1]
	if p1 == p2 { // no items for new state. skip. (just safeguard)
		return
	}
	// sort to help search later
	sort.Slice(kernls[p1:p2], func(i, j int) bool {
		return kernls[i].prd.id < kernls[j].prd.id ||
			(kernls[i].prd.id == kernls[j].prd.id && kernls[i].off < kernls[j].off)
	})

	if prev := searchState(nstate, sym); prev > 0 { // found previous state, merge
		q1, q2 := kernlp[prev], kernlp[prev+1]
		for k, l := p1, q1; l < q2; k, l = k+1, l+1 {
			kernls[l].lkset.union(kernls[k].lkset)
		}
		return
	}
	// not found, completely new state!
	if sym >= NTBASE {
		stateChain[nstate] = ntstates[sym-NTBASE]
		ntstates[sym-NTBASE] = nstate
	} else {
		stateChain[nstate] = ntstates[sym]
		tstates[sym] = nstate
	}
	nstate++
}

// searchState looks for the identical previous state of the same transition symbol (without considering the lookahead set)
func searchState(n int, sym int) int {
	p1, p2 := kernlp[n], kernlp[n+1]
	var prev int
	if sym >= NTBASE {
		prev = ntstates[sym-NTBASE]
	} else {
		prev = tstates[sym]
	}
prevState:
	for ; prev != 0; prev = stateChain[prev] {
		q1, q2 := kernlp[prev], kernlp[prev+1]
		if (p2 - p1) != (q2 - q1) { // kernels are not the same size
			continue
		}
		for k, l := p1, q1; l < q2; k, l = k+1, l+1 {
			if kernls[k].prd.id != kernls[l].prd.id || kernls[k].off != kernls[l].off {
				continue prevState
			}
		}
		return prev
	}

	return -1
}