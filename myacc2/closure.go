package myacc2

const NTBASE = 010000
const WSETINC = 50 // increase for working sets    wsets
const NSTATES = 16000

type prd struct {
	id  int
	prd []int
}

type item struct {
	off   int   // position of symbol right after dot
	prd   prd   // pointer into prds
	first int   // symbol right after dot
	lkset lkset // lookahead set
}

func (i item) clone() item {
	// lkset will be manipulated in closure. clone it
	return item{i.off, i.prd, i.first, i.lkset.clone()}
}

func (i item) string() string {
	return ""
}

type wItem struct {
	item      item
	processed bool // mark processed during closure
	done      bool // mark done for state generation
}

// type state struct {
// 	kernel []item // kernel items
// 	// partitions map[int][]
// }

var cwp int
var wSet []wItem                    // store temporary items generated during closure
var kernlp = make([]int, NSTATES+2) // pointer to kernel items
var kernls []item
var clkset lkset
var tbitset = 0           // size of lookahead sets
var firstSets []lkset     // the first sets of symbols
var empty []bool          // to check whether a symbol is nullable
var prdsStartWith [][]prd // 1st dimension is the symbol, 2nd dimension is prds started with it
var allPrds []prd

/* closure computes closure for state n*/
func closure(n int) {
	cwp = 0 // reset current pointer for working set
	// copy kernel items of new state
	for p := kernlp[n]; p < kernlp[n+1]; p++ {
		wSet[cwp].item = kernls[p].clone()
		wSet[cwp].processed, wSet[cwp].done = false, false
		cwp++
	}
	fill(clkset, tbitset, 0)
	/* changed reflects that
	- an item is added to closure
	- lookahead set of an item is changed (by merged)
	these can:
		- add more items into closure
		- change lookahead sets of items in closure
	*/
	changed := true
	for changed {
		changed = false
		for i := 0; i < cwp; i++ {
			if wSet[i].processed {
				continue
			}
			wSet[i].processed = true // mark processed, skip next time

			first := wSet[i].item.first
			if first < NTBASE { // terminal symbol or action, skip
				continue
			}
			// non-terminal symbol, get lookahead set for items derived from this item
			itemI := wSet[i].item
			prdI := itemI.prd
			j := itemI.off + 1
			sym := prdI.prd[j] // symbol right after first
			for sym > 0 {
				if sym < NTBASE {
					clkset.set(sym) // terminal symbol, set it and stop
					break
				}
				// non-terminal symbol, merge its first set
				clkset.union(firstSets[sym-NTBASE])
				if !empty[sym-NTBASE] { // not nullable, stop
					break
				}
				j++
				sym = prdI.prd[j]
			}
			if sym <= 0 { // reach the end
				clkset.union(itemI.lkset)
			}
			// the productions that has first as LHS
			prds := prdsStartWith[first-NTBASE]

		nextPrd:
			for _, prdI = range prds {
				/* if an item is already derived with prdI,
				for example, an item derived itself! t -> .t*f
				merge its lookahead set with clkset*/
				for i := 0; i < cwp; i++ {
					itemI := wSet[i].item
					// derived item always has oft = 1
					if itemI.off == 1 && itemI.prd.id == prdI.id {
						if itemI.lkset.union(clkset) {
							changed = true
							wSet[i].processed = false // mark to process later. may change closure of derived items
						}
						continue nextPrd

					}
				}

				if cwp > len(wSet) {
					extend(&wSet, WSETINC)
				}
				wSet[cwp].item = item{1, prdI, prdI.prd[1], clkset.clone()}
				wSet[cwp].processed, wSet[cwp].done = false, false
				changed = true
				cwp++
			}
		}
	}
}
