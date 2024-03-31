package myacc2

import "fmt"

// handle shift/reduce conflict in state s, between rule r and terminal symbol t
func shiftReduceConflict(r, t, s int) {
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
	as := LASC // reduce
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
func cpReds(i int) {
	red := 0
	for _, w := range wSet[:cwp] {
		first := w.item.first
		if first > 0 {
			continue
		}
		red = -first
		lk := w.item.lkset
		for j, s := range trans[:termN] {
			if !lk.check(s) {
				continue
			}
			if s == 0 {
				trans[j] = red
			} else if s < 0 { // reduce/reduce conflict
				fmt.Fprintf(foutput, "\n %v: reduce/reduce conflict  (red'ns %v and %v) on %v",
					i, -s, red, terms[j].name)
				if -s > red { // favor rule higher in grammar
					trans[j] = red
				}
			} else { // shift/reduce conflict
				shiftReduceConflict(red, j, i)
			}
		}
	}
	// compute reduction with maximum count
	maxRed = 0
	maxCount := 0
	redCounts := map[int]int{}
	for _, act := range trans[:termN] {
		if act < 0 {
			r := -act
			redCounts[r]++
			if redCounts[r] > maxCount {
				maxCount = redCounts[r]
				maxRed = r
			}
		}
	}
	for i, act := range trans[:termN] { // clear default reduction cells
		if act == -maxRed {
			trans[i] = 0
		}
	}
	defReds[i] = maxRed
}

func writeState(i int) {
	fmt.Fprintf(foutput, "\nstate %v\n", i)
	for _, a := range kernls[kernlp[i]:kernlp[i+1]] {
		fmt.Fprintf(foutput, "\t%s\n", a.string())
	}
	if hasEpsilons[i] {
		for _, w := range wSet[kernlp[i+1]-kernlp[i] : cwp] {
			if w.item.first < 0 {
				fmt.Fprintf(foutput, "\t%s\n", w.item.string())
			}
		}
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
		if hasEpsilons[i] {
			closure(i)
		} else {
			closure0(i)
		}
		fill(trans, termN+nontermN, 0)
		for j, w := range wSet[:cwp] {
			first := w.item.first
			if first > 1 && first < NTBASE && trans[first] == 0 {
				for _, v := range wSet[j:cwp] {
					if first == v.item.first {
						addKernItem(v.item)
					}
				}
				trans[first] = retrieveState(first)
			} else if first > NTBASE {
				first -= NTBASE
				trans[first+termN] = goTos[gotoIdx[i]+first]
			}
		}
		if i == 1 {
			trans[1] = ACCEPTCODE
		}
		cpReds(i)
		writeState(i)
	}

}
