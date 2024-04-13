package myacc2

import "fmt"

// ======== must be initialized when nstate is known!
var shifts [][]int // first dimension is state, second dimension is pairs of (terminal symbol, next state)

// handle shift/reduce conflict in state s, between rule r and terminal symbol t
func srConflict(r, t, s int) {
	lp := aptPrd[r]
	lt := aptTerm[t]
	if prec(lp) == 0 || prec(lt) == 0 { // shift
		fmt.Fprintf(foutput, "\n%d: shift/reduce conflict (shift %d(%d), red'n %d(%d)) on %s",
			s, trans[t], prec(lt), r, prec(lp), terms[t].name)
		return
	}
	if prec(lt) > prec(lp) { // shift
		return
	}
	as := LASC // check associativity
	if prec(lt) == prec(lp) {
		as = asc(lt)
	}
	switch as {
	case BASC:
		trans[t] = ERRCODE
	case LASC: // reduce
		trans[t] = -r
	}
	// shift in other cases
}

// compute reductions after closing a state
func cReds(i int) {
	red := 0

	// check if an item induces reduction
	// handle reduce/reduce conflict or shift/reduce conflict
	cRedItem := func(itemI item) {
		first := itemI.first
		if first > 0 {
			return
		}
		red = -first
		lk := itemI.lkset
		for j, s := range trans[:termN] {
			if !lk.check(s) {
				continue
			}
			if s == 0 {
				trans[j] = red
			} else if s < 0 { // reduce/reduce conflict. don't need to check the case s == red!
				fmt.Fprintf(foutput, "\n %v: reduce/reduce conflict  (red'ns %v and %v) on %v",
					i, -s, red, terms[j].name)
				if -s > red { // favor rule higher in grammar
					trans[j] = red
				}
			} else { // shift/reduce conflict
				srConflict(red, j, i)
			}
		}
	}
	// only consider kernel and epsilon items for the reduction of state i
	for _, itemI := range kernls[kernlp[i]:kernlp[i+1]] {
		cRedItem(itemI)
	}
	for _, itemI := range epsilons[i] {
		cRedItem(itemI)
	}
	// compute reduction with maximum count
	maxRed = 0
	maxCount := 0
	redCounts := map[int]int{}
	for _, act := range trans[:termN] {
		if act < 0 {
			r := -act
			aptPrd[r] |= REDFLAG // track if a prd ever reduces
			redCounts[r]++
			if redCounts[r] > maxCount {
				maxCount = redCounts[r]
				maxRed = r
			}
		}
	}
	for i, act := range trans[:termN] { // clear default reduction cells. the cell with 0s will also be treated as default reduction later!!!
		if act == -maxRed {
			trans[i] = 0
		}
	}
	defReds[i] = maxRed
}

// compute shifts after closing state i
func cShifts(i int) {
	n := 0
	// counting the shifts
	for _, act := range trans[:termN+1] {
		if act > 0 && act != ACCEPTCODE && act != ERRCODE {
			n++
		}
	}
	row := make([]int, n*2) // pairs of (terminal symbol, next state)
	n = 0
	// this is repetitive, but more efficient for memory allocation
	for t, act := range trans[:termN+1] {
		if act > 0 && act != ACCEPTCODE && act != ERRCODE {
			row[n] = t
			n++
			row[n] = act
			n++
		}
	}
	shifts[i] = row
}

func writeState(i int) {
	fmt.Fprintf(foutput, "\nstate %v\n", i)
	// write kernels
	for _, a := range kernls[kernlp[i]:kernlp[i+1]] {
		fmt.Fprintf(foutput, "\t%s\n", a.string())
	}
	// write epsilons
	for _, itemI := range epsilons[i] {
		fmt.Fprintf(foutput, "\t%s\n", itemI.string())
	}

	for t, s := range trans[:termN] {
		if s > 0 {
			fmt.Fprintf(foutput, "\n\t%s  ", terms[t].name) //======= implement later!!!
			switch s {
			case ACCEPTCODE:
				fmt.Fprintf(foutput, "accept")
			case ERRCODE:
				fmt.Fprintf(foutput, "error")
			default: // s > 0
				fmt.Fprintf(foutput, "shift %v", s)
			}
		} else if s < 0 { // other remaining reductions (not default)
			fmt.Fprintf(foutput, "reduce %d (src line %d)", -s, -11111)
		}
	}
	if maxRed != 0 {
		fmt.Fprintf(foutput, "\n\t.  reduce %d (src line %d)\n\n", maxRed, -11111)
	} else {
		fmt.Fprintf(foutput, "\n\t.  error\n\n")
	}
	for n, s := range trans[termN+1:] {
		if s != 0 {
			fmt.Fprintf(foutput, "\t%s  goto %d\n", nonterms[n].name, s)
		}
	}
}

func output() {
	for i := 0; i < nstate; i++ {
		closure0(i)
		fill(trans, termN+nontermN, 0)
		for j, w := range wSet[:cwp] {
			first := w.item.first
			if first > 1 && first < NTBASE && trans[first] == 0 {
				for _, v := range wSet[j:cwp] {
					if first == v.item.first {
						addKernItem(v.item)
					}
				}
				trans[first] = retrieveState(first) // a shift
			} else if first > NTBASE {
				first -= NTBASE
				trans[first+termN] = actionStore[gotoIdx[i]+first] // a goto
			}
		}
		if i == 1 {
			trans[1] = ACCEPTCODE
		}
		cReds(i)
		cShifts(i)
		writeState(i)
	}

}
