package myacc2

import "fmt"

var allPrds []prd     // holds all production rules declared in input
var prdYields [][]prd // 1st dimension is the symbol, 2nd dimension is prds started with it
var empty []bool      // to check whether a symbol is nullable
var firstSets []lkset // the first sets for each symbol

// production has an id and a list of symbols
type prd struct {
	id  int
	prd []int
}

// compute prdYiedls for each non-terminal symbol
func cPrdYields() {
	prdYields = make([][]prd, nontermN+1)

	yield := make([]prd, nprod)

	if false {
		for j, t := range nonterms[:nontermN+1] {
			fmt.Printf("nonterms[%d] = %s\n", j, t.name)
		}
		for j, prod := range allPrds {
			fmt.Printf("allPrds[%d][0] = %d+NTBASE\n", j, prod.prd[0]-NTBASE)
		}
	}

	fatfl = 0
	for i, s := range nonterms[:nontermN+1] {
		n := 0
		c := i + NTBASE
		for _, prod := range allPrds {
			if prod.prd[0] == c {
				yield[n] = prod
				n++
			}
		}
		if n == 0 {
			errorf("nonterminal %s doesn't yield any production", s.name)
			continue
		}
		prdYields[i] = make([]prd, n)
		copy(prdYields[i], yield)
	}
	fatfl = 1
	if nerrors != 0 {
		// TODO: summary
		// exit(1)
	}
}

func cEmpty() {
	through := make([]bool, nontermN+1) // a symbol can derive through the end if all symbols on its RHS can derive through the end

thru: // check if any symbol can not derive through the end
	for {
		for _, prod := range allPrds {
			if through[prod.prd[0]-NTBASE] { // already checked
				continue
			}
			i := 0
			for _, s := range prod.prd[1 : len(prod.prd)-1] {
				if s >= NTBASE && !through[s-NTBASE] {
					break
				}
				i++
			}
			if i == len(prod.prd)-1 {
				through[prod.prd[0]-NTBASE] = true
				continue thru
			}
		}
		break
	}

	for i, s := range nonterms[1 : nontermN+1] {
		if !through[i] {
			fatfl = 0
			errorf("nonterminal " + s.name + " can not derive through the end")
			fatfl = 1
		}
	}
	if nerrors != 0 {
		//TODO summary, exit()
	}
	empty = make([]bool, nontermN+1)

search:
	for {
	nextPrd:
		for _, prod := range allPrds[1:] {
			if empty[prod.prd[0]] { // already checked
				continue
			}

			for _, s := range prod.prd[1:] {
				if s < NTBASE || !empty[s-NTBASE] {
					continue nextPrd
				}
			}
			empty[prod.prd[0]-NTBASE] = true
			// discover a new empty non-term. search all over again from first prd
			continue search
		}
		return // done searching
	}
}

// compute firstset for each non-terminal symbol
func cFirstSets() {
	firstSets = make([]lkset, nontermN+1)
	var yield []prd
	// fill firstset with terminal appears first on RHS
	for i := range nonterms[:nontermN+1] { // seems don't need end idx here!
		firstSets[i] = newLkset()
		yield = prdYields[i] // prds started with s

		for _, prod := range yield {
			for _, ch := range prod.prd[1 : len(prod.prd)-1] {
				if ch < NTBASE { // terminal symbol, set it in firstset, move on to next prd
					firstSets[i].set(ch)
					break
				}
				if !empty[ch-NTBASE] { // non-nullable non-terminal symbol. move on to next prd
					break
				}
			}
		}
	}

	// transitivity: if a non-term on RHS can appear first, its firstset will be unioned with LHS's
	changed := true // changed reflects if firstset of any non-term changed through transitivity. if so, search the whole prd list again
	for !changed {
		changed = false
		for i := range nonterms[:nontermN+1] {
			yield = prdYields[i]
			for _, prod := range yield {
				for _, ch := range prod.prd[1 : len(prod.prd)-1] {
					if ch < NTBASE { // terminal symbol appears first. break
						break
					}
					ch -= NTBASE
					changed = changed || firstSets[i].union(firstSets[ch])
					if !empty[ch] {
						break
					}
				}
			}
		}
	}
}
