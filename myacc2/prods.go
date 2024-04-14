package myacc2

import "fmt"

var prods [][]int     // holds all production rules declared in input
var yields [][][]int  // 1st dimension is the symbol, 2nd dimension is prds started with it
var empty []bool      // to check whether a symbol is nullable
var firstSets []lkset // the first sets for each symbol

// get id of a production
func id(prd []int) int {
	return -prd[len(prd)-1]
}

// compute production-yield for each non-terminal symbol
func cYields() {
	yields = make([][][]int, nontermN+1)

	yld := make([][]int, nprod)

	if false {
		for j, t := range nonterms[:nontermN+1] {
			fmt.Printf("nonterms[%d] = %s\n", j, t.name)
		}
		for j, prd := range prods {
			fmt.Printf("allPrds[%d][0] = %d+NTBASE\n", j, prd[0]-NTBASE)
		}
	}

	fatfl = 0
	for i, s := range nonterms[:nontermN+1] {
		n := 0
		c := i + NTBASE
		for _, prd := range prods {
			if prd[0] == c {
				yld[n] = prd
				n++
			}
		}
		if n == 0 {
			errorf("nonterminal %s doesn't yield any production", s.name)
			continue
		}
		yields[i] = make([][]int, n)
		copy(yields[i], yld)
	}
	fatfl = 1
	if nerrors != 0 {
		// TODO: summary
		// exit(1)
	}
}

// to check if a non-term symbol is nullable. use in computing firstsets
func cEmpty() {
	through := make([]bool, nontermN+1) // a symbol can derive through the end if all symbols on its RHS can derive through the end

thru: // check if any symbol can not derive through the end
	for {
		for _, prd := range prods {
			if through[prd[0]-NTBASE] { // already checked
				continue
			}
			i := 0
			for _, s := range prd[1 : len(prd)-1] {
				if s >= NTBASE && !through[s-NTBASE] {
					break
				}
				i++
			}
			if i == len(prd)-1 {
				through[prd[0]-NTBASE] = true
				continue thru // discover new symbol. search all over again
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
	empty = make([]bool, nontermN+1) // a LHS symbol is empty if there's a rule of its yield that RHS is ɛ or all the RHS symbol is empty

search:
	for {
	nextPrd:
		for _, prd := range prods[1:] {
			if empty[prd[0]] { // already checked
				continue
			}

			for _, s := range prd[1:] {
				if s < NTBASE || !empty[s-NTBASE] {
					continue nextPrd
				}
			}
			empty[prd[0]-NTBASE] = true
			// discover a new empty non-term. search all over again from first prd
			continue search
		}
		return // done searching
	}
}

// compute firstset for each non-terminal symbol
func cFirstSets() {
	firstSets = make([]lkset, nontermN+1)
	var yield [][]int
	// fill firstset with terminal appears first on RHS
	for i := range nonterms[:nontermN+1] { // seems don't need end idx here!
		firstSets[i] = newLkset()
		yield = yields[i] // prds started with s

		for _, prd := range yield {
			for _, ch := range prd[1 : len(prd)-1] {
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
			yield = yields[i]
			for _, prd := range yield {
				for _, ch := range prd[1 : len(prd)-1] {
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
