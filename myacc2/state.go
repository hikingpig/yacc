package myacc2

import (
	"fmt"
	"sort"
)

var nstate int = 0

const STATEINC = 200

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

func stategen() {
	nstate = 0
	clkset = newLkset()
	addKernItem(item{0, allPrds[0], 0, newLkset()})
	nstate++
	for {
		mNstate := nstate
		// if closure doesn' add any items and doesn't change existing kernel items?
		// it still can generate new states if first symbol of some items is not at the end (<=0)
		// when these items are put in kernels, the dot move 1 notch to the right. that's the kernel of new state
		printState()
		closure(nstate - 1)
		fmt.Printf("======= wset: %+v\n", wSet[:5])
		for i := 0; i < cwp; i++ {
			w := wSet[i]
			fmt.Printf("===== item %d, %+v\n", i, w)
			if w.done {
				continue
			}
			w.done = true
			first := w.item.first
			if first <= 0 { // end of prd. skip
				continue
			}
			kernlp[nstate+1] = kernlp[nstate] // next state hold 0 items initially
			addKernItem(w.item)
			for j := i + 1; j < cwp; j++ {
				if first == wSet[j].item.first {
					addKernItem(wSet[j].item)
					wSet[j].done = true
				}
			}
			sort.Slice(kernls[kernlp[nstate]:kernlp[nstate+1]], func(i, j int) bool {
				return kernls[i].prd.id < kernls[j].prd.id ||
					(kernls[i].prd.id == kernls[j].prd.id && kernls[i].off < kernls[j].off)
			})
			nstate++
		}
		if nstate == mNstate {
			break
		}
	}
}
