package myacc2

import (
	"bufio"
	"sort"
)

const STATEINC = 200
const TEMPSIZE = 16000
const ACTSIZE = 240000
const ACCEPTCODE = 8191
const ERRCODE = 8190

// no, left, right, binary assoc.
const (
	NOASC = iota
	LASC
	RASC
	BASC
)

var gotoIdx = make([]int, NSTATES)
var goTos = make([]int, ACTSIZE)
var lastGoto int // last goto's index in goTos

var nstate = 0
var stateChain = make([]int, NSTATES) // chain states of the same transition symbol together
var tstates []int                     // latest state generated by a terminal transition symbol
var ntstates []int                    // latest state generated by a non-terminal transition symbol
var termN = 0                         // idx of terminal symbol in terms. starting from 1
var nontermN = -1                     // idx of non-terminal symbol in non-terms, including $accept. must add 1 to get total number of nonterms!
var trans []int                       // states generated by the transition symbols during closure of an existing state
var foutput *bufio.Writer             // file to write state output
var maxRed int                        // default reduction of a state
var defReds []int                     // default reductions of states
var aptPrd []int                      // compact info: associativity, precedence, type of prds
var aptTerm []int                     // compact info: associativity, precedence, type of terminal symbols
var hasEpsilons = map[int]bool{}

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
	tstates = make([]int, termN)     // include $end(0)
	ntstates = make([]int, nontermN) // include $accept(NTBASE)
	trans = make([]int, nontermN)
	clkset = newLkset()
	addKernItem(item{0, allPrds[0], 0, newLkset()})
	nstate++
	for i := 0; i < nstate; i++ {
		closure(i)
		for j := 0; j < cwp; j++ {
			w := wSet[j]
			if w.done {
				continue
			}
			w.done = true
			first := w.item.first
			// ============ 0 is for prd 0!, $end should be from 1
			// ============= should we skip 0???
			if first == 0 || first == NTBASE { // skip $end and $accept
				continue
			}
			if first < 0 { // end of prd after derived
				if j >= kernlp[i+1]-kernlp[i] {
					hasEpsilons[i] = true // ========= in output: do closure if has epsilon, closure0 if no epsilon
				}
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
			s := newState(first)
			if first > NTBASE { // terminal symbol
				trans[first-NTBASE] = s
			}
		}
		gotoIdx[i] = packGotos(trans[termN:], nontermN-1) - 1 // $accept is excluded from goTos
	}
}

// newState creates a new state for a transtion symbol or merge it with existing state
func newState(sym int) int {
	p1, p2 := kernlp[nstate], kernlp[nstate+1]
	if p1 == p2 { // no items for new state. skip. (just safeguard)
		return 0
	}
	// sort to help search later
	sort.Slice(kernls[p1:p2], func(i, j int) bool {
		return kernls[i].prd.id < kernls[j].prd.id ||
			(kernls[i].prd.id == kernls[j].prd.id && kernls[i].off < kernls[j].off)
	})

	if prev := searchState(nstate, sym); prev > 0 { // found previous state, merge
		q1, q2 := kernlp[prev], kernlp[prev+1]
		kernlp[nstate+1] = kernlp[nstate] // delete current items
		for k, l := p1, q1; l < q2; k, l = k+1, l+1 {
			kernls[l].lkset.union(kernls[k].lkset)
		}
		return prev
	}
	// not found, completely new state!
	if sym >= NTBASE {
		stateChain[nstate] = ntstates[sym-NTBASE]
		ntstates[sym-NTBASE] = nstate
	} else {
		stateChain[nstate] = tstates[sym]
		tstates[sym] = nstate
	}
	nstate++
	return nstate - 1
}

// retrieve existing state for current items
func retrieveState(sym int) int {
	p1, p2 := kernlp[nstate], kernlp[nstate+1]
	kernlp[nstate+1] = kernlp[nstate] // delete current items
	if p1 == p2 {                     // no items for new state. skip. (just safeguard)
		return 0
	}
	// sort to help search later
	sort.Slice(kernls[p1:p2], func(i, j int) bool {
		return kernls[i].prd.id < kernls[j].prd.id ||
			(kernls[i].prd.id == kernls[j].prd.id && kernls[i].off < kernls[j].off)
	})

	if prev := searchState(nstate, sym); prev > 0 { // found previous state, merge
		return prev
	}
	return 0
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

// compact goTos by removing leading and trailing zeros and duplicated parts
func packGotos(trans []int, n int) int {
	m := 0
	// skip leading zeros
	for m < n && trans[m] == 0 {
		m++
	}
	if m >= n {
		return 0
	}
	// skip trailing zeros
	for n > m && trans[n] == 0 {
		n--
	}
	trans = trans[m : n+1]
	maxStart := len(goTos) - len(trans)
nextStart:
	for start := 0; start < maxStart; start++ {
		// find spot with enough empty room for new goTos
		for i, j := 0, start; i < len(trans); i, j = i+1, j+1 {
			if trans[i] != 0 && goTos[j] != 0 && trans[i] != goTos[j] {
				continue nextStart
			}
		}
		// found it, store action
		for i, j := 0, start; i < len(trans); i, j = i+1, j+1 {
			if trans[i] != 0 {
				goTos[j] = trans[i]
				if j > lastGoto {
					lastGoto = j
				}
			}
		}

		return -m + start
	}
	// can not find empty spots for new goTos
	return 0
}
