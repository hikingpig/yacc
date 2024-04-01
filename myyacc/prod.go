package myyacc

import "fmt"

var pres [][][]int // see cpres
var pempty []int   // mark nullable non-terminal symbols
var pfirst []Lkset // lookahead sets for non-terminal symbols
var tbitset = 0    // size of lookahead sets
var indebug = 0    // debugging flag for cpfir

/*
cpres computes pres:
- the first dimension is LHS non-terminal symbol,
- the second dimension is production number,
- the third is symbols in that rule
the first item in the rule is omitted as it is the symbol itself!
*/
func cpres() {
	pres = make([][][]int, nterm+1)
	curres := make([][]int, nprod)

	if false {
		for j := 0; j <= nnonterm; j++ {
			fmt.Printf("nnonter[%v] = %v\n", j, nonterms[j].name)
		}
		for j := 0; j < nprod; j++ {
			fmt.Printf("prdptr[%v][0] = %v+NTBASE\n", j, prdptr[j][0]-NTBASE)
		}
	}

	fatfl = 0 // make undefined symbols nonfatal
	for i := 0; i <= nnonterm; i++ {
		n := 0
		c := i + NTBASE
		for j := 0; j < nprod; j++ {
			if prdptr[j][0] == c { // ========= should use map for more efficient access? NO!
				curres[n] = prdptr[j][1:]
				n++
			}
		}
		if n == 0 {
			errorf("nonterminal %v not defined", nonterms[i].name)
			continue
		}
		pres[i] = make([][]int, n)
		copy(pres[i], curres)
	}
	fatfl = 1
	if nerrors != 0 {
		summary()
		exit(1)
	}
}

/*
cpres computes pres:
- the first dimension is LHS non-terminal symbol,
- the second dimension is production number,
- the third is symbols in that rule
* don't skip the LHS. keep the rule the same as that in prdptr
*/
func cpres2() {
	pres = make([][][]int, nterm+1)
	curres := make([][]int, nprod)

	if false {
		for j := 0; j <= nnonterm; j++ {
			fmt.Printf("nnonter[%v] = %v\n", j, nonterms[j].name)
		}
		for j := 0; j < nprod; j++ {
			fmt.Printf("prdptr[%v][0] = %v+NTBASE\n", j, prdptr[j][0]-NTBASE)
		}
	}

	fatfl = 0 // make undefined symbols nonfatal
	for i := 0; i <= nnonterm; i++ {
		n := 0
		c := i + NTBASE
		for j := 0; j < nprod; j++ {
			if prdptr[j][0] == c { // ========= should use map for more efficient access? NO!
				curres[n] = prdptr[j] // ============= keep preference, don't copy it over!
				n++
			}
		}
		if n == 0 {
			errorf("nonterminal %v not defined", nonterms[i].name)
			continue
		}
		pres[i] = make([][]int, n)
		copy(pres[i], curres)
	}
	fatfl = 1
	if nerrors != 0 {
		summary()
		exit(1)
	}
}

func cempty() {
	var i, p, np int
	var prd []int

	pempty = make([]int, nterm+1) // =========== naming it like this is misleading, it should be pOK!!

	// first, use the array pempty to detect productions that can never be reduced
	// set pempty to WHONOWS
	aryfil(pempty, nnonterm+1, WHOKNOWS)

	// now, look at productions, marking nonterminals which derive something
	// ============= it seems to check that all non-terminal symbols on the right must derive something
	// ============= not at least one of them
more:
	for {
		for i = 0; i < nprod; i++ {
			prd = prdptr[i]
			if pempty[prd[0]-NTBASE] != 0 { // ======== use WHOKNOWS
				continue
			}
			np = len(prd) - 1
			for p = 1; p < np; p++ { // ========== why skip last symbol?// it is 0
				if prd[p] >= NTBASE && pempty[prd[p]-NTBASE] == WHOKNOWS { // its non-terminal, and it is not determined, stop
					break
				} // ============= prd[p]-NTBASE?? what if it is negative??
			}
			// production can be derived
			if p == np { // all symbols on the right handside are determined!!! // IT IS NOT AT LEAST one of them?
				pempty[prd[0]-NTBASE] = OK
				continue more
			}
		}
		break
	}

	// now, look at the nonterminals, to see if they are all OK
	for i = 0; i <= nnonterm; i++ {
		// the added production rises or falls as the start symbol ...
		if i == 0 {
			continue
		}
		if pempty[i] != OK {
			fatfl = 0
			errorf("nonterminal " + nonterms[i].name + " never derives any token string")
		}
	}

	if nerrors != 0 {
		summary()
		exit(1)
	}

	// now, compute the pempty array, to see which nonterminals derive the empty string
	// set pempty to WHOKNOWS
	aryfil(pempty, nnonterm+1, WHOKNOWS)

	// loop as long as we keep finding empty nonterminals

	// ========= EMPTY means not deriving anything, not empty string
again:
	for {
	next:
		for i = 1; i < nprod; i++ {
			// not known to be empty
			prd = prdptr[i]
			if pempty[prd[0]-NTBASE] != WHOKNOWS {
				continue
			}
			//======= check righthandside symbols
			// ======== there should be no terminal symbols and all non-terminal symbol are empty
			np = len(prd) - 1
			for p = 1; p < np; p++ {
				if prd[p] < NTBASE || pempty[prd[p]-NTBASE] != EMPTY {
					continue next
				}
			}
			// the last symbol is [..., 0], indicating empty string
			// we have a nontrivially empty nonterminal
			pempty[prd[0]-NTBASE] = EMPTY

			// got one ... try for another
			continue again
		}
		return
	}
}

// compute an array with the first of nonterminals
func cpfir() {
	var s, n, p, np, ch, i int
	var curres [][]int
	var prd []int

	wsets = make([]Wset, nnonterm+WSETINC)
	pfirst = make([]Lkset, nnonterm+1)
	for i = 0; i <= nnonterm; i++ { // ============ why it is == nnonter? we have nnonter+1 symbols?
		wsets[i].ws = newLkset() // the role of wsets is not show here!!!! maybe used where else.
		pfirst[i] = newLkset()
		curres = pres[i]
		n = len(curres)

		// initially fill the sets
		for s = 0; s < n; s++ {
			prd = curres[s]
			np = len(prd) - 1
			for p = 0; p < np; p++ {
				ch = prd[p]
				if ch < NTBASE {
					pfirst[i].set(ch) // set the bit ch in pfirst[i], each Lkset holds all the terminal tokens
					break             // move on to next production rule, got a terminal symbol here!
				}
				if pempty[ch-NTBASE] == 0 { // it's a non-terminal and not empty. move to next production rule
					break // do we need to keep it for reference?
				}
			}
		}
	}

	// now, reflect transitivity
	for { // ========== assume that symbol i changed, but symbol i+1 doesn't change. so changes is 0. but maybe symbol i-1 depends on symbol i, but we terminate the loop so symbol i-1 will not be updated

		changed := false
		for i = 0; i <= nnonterm; i++ {
			curres = pres[i]
			n = len(curres)
			for s = 0; s < n; s++ {
				prd = curres[s]
				np = len(prd) - 1
				for p = 0; p < np; p++ {
					ch = prd[p] - NTBASE
					if ch < 0 { // terminal symbol, stop, move to next rule
						break
					}
					if pfirst[i].union(pfirst[ch]) {
						changed = true
					}
					// =========== instead of checking changes outside, check it here
					// if changes != 0 { continue out}
					if pempty[ch] == 0 { // non-terminal, not empty string, stop
						break
					}
				}
			}
		}
		if !changed {
			break
		}
	}

	if indebug == 0 {
		return
	}
	if foutput != nil {
		for i = 0; i <= nnonterm; i++ {
			fmt.Fprintf(foutput, "\n%v: %v %v\n",
				nonterms[i].name, pfirst[i], pempty[i])
		}
	}
}
