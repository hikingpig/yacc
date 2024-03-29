package myyacc

import "fmt"

// flags for state generation
const (
	DONE = iota
	MUSTDO
	MUSTLOOKAHEAD
)

type Wset struct {
	pitem Pitem
	flag  int
	ws    Lkset
}

var nstate = 0                      // number of states
var pstate = make([]int, NSTATES+2) // items of state i = statemem[pstate[i]:pstate[i+1]]
var tystate = make([]int, NSTATES)  // contains type information about the states
var tstates []int                   // states generated by terminal gotos
var ntstates []int                  // states generated by nonterminal gotos
var mstates = make([]int, NSTATES)  // chain of overflows of term/nonterm generation lists

var amem []int                   // action table storage
var memp int                     // next free action table position
var indgo = make([]int, NSTATES) // index to the stored goto table

var maxspr int // maximum spread of any entry
var maxoff int // maximum offset into a array
var maxa int

type Row struct {
	actions       []int
	defaultAction int
}

var stateTable []Row

var zznewstate = 0

// statistics collection variables

var zzgoent = 0
var zzgobest = 0
var zzacent = 0
var zzexcp = 0
var zzclose = 0
var zzrrconf = 0
var zzsrconf = 0
var zzstate = 0

var wsets []Wset
var cwp int

var clset Lkset                   // temporary storage for lookahead computations
var temp1 = make([]int, TEMPSIZE) // temporary storage, indexed by terms + ntokens or states
var gsdebug = 0                   // debugging flag for stagen
var cldebug = 0                   // debugging flag for closure
var pkdebug = 0                   // debugging flag for apack

func stagen() {
	// initialize
	nstate = 0
	tstates = make([]int, nterm+1)     // states generated by terminal gotos
	ntstates = make([]int, nnonterm+1) // states generated by nonterminal gotos
	amem = make([]int, ACTSIZE)        // memory allocation
	memp = 0

	clset = newLkset()
	pstate[0] = 0
	pstate[1] = 0
	aryfil(clset, tbitset, 0)
	putitem(Pitem{prdptr[0], 0, 0, 0}, clset)
	tystate[0] = MUSTDO
	nstate = 1
	pstate[2] = pstate[1]

	//
	// now, the main state generation loop
	// first pass generates all of the states
	// later passes fix up lookahead
	// could be sped up a lot by remembering
	// results of the first pass rather than recomputing
	//
	first := 1
	for more := 1; more != 0; first = 0 {
		more = 0
		for i := 0; i < nstate; i++ {
			if tystate[i] != MUSTDO {
				continue
			}

			tystate[i] = DONE
			aryfil(temp1, nnonterm+1, 0)

			// take state i, close it, and do gotos
			closure(i)

			// generate goto's
			for p := 0; p < cwp; p++ {
				pi := wsets[p]
				if pi.flag != 0 {
					continue
				}
				wsets[p].flag = 1
				c := pi.pitem.first
				if c <= 1 {
					if pstate[i+1]-pstate[i] <= p {
						tystate[i] = MUSTLOOKAHEAD
					}
					continue
				}

				// do a goto on c
				putitem(wsets[p].pitem, wsets[p].ws)
				for q := p + 1; q < cwp; q++ {
					// this item contributes to the goto
					if c == wsets[q].pitem.first {
						putitem(wsets[q].pitem, wsets[q].ws)
						wsets[q].flag = 1
					}
				}

				if c < NTBASE {
					state(c) // register new state
				} else {
					temp1[c-NTBASE] = state(c)
				}
			}

			if gsdebug != 0 && foutput != nil {
				fmt.Fprintf(foutput, "%v: ", i)
				for j := 0; j <= nnonterm; j++ {
					if temp1[j] != 0 {
						fmt.Fprintf(foutput, "%v %v,", nonterms[j].name, temp1[j])
					}
				}
				fmt.Fprintf(foutput, "\n")
			}

			if first != 0 {
				indgo[i] = apack(temp1[1:], nnonterm-1) - 1
			}

			more++
		}
	}
}

// generate the closure of state i
func closure(i int) {
	zzclose++
	fmt.Printf("=========== i = %d, zzclose = %d\n", i, zzclose)
	// first, copy kernel of state i to wsets
	cwp = 0
	fmt.Printf("=========== pstate[i] = %d, pstate[i+1] = %d\n", pstate[i], pstate[i+1])
	for p := pstate[i]; p < pstate[i+1]; p++ {
		wsets[cwp].pitem = statemem[p].pitem
		wsets[cwp].flag = 1 // this item must get closed
		copy(wsets[cwp].ws, statemem[p].look)
		// fmt.Printf("======= cwp = %d, wsets[cwp] = %+v\n", cwp, wsets[cwp])
		cwp++
	}

	// now, go through the loop, closing each item
	work := 1
	for work != 0 {
		work = 0
		for u := 0; u < cwp; u++ {
			if wsets[u].flag == 0 {
				continue
			}

			// dot is before c
			c := wsets[u].pitem.first
			if c < NTBASE {
				wsets[u].flag = 0
				// only interesting case is where . is before nonterminal
				continue
			}

			// compute the lookahead
			aryfil(clset, tbitset, 0)

			// find items involving c
			for v := u; v < cwp; v++ {
				if wsets[v].flag != 1 || wsets[v].pitem.first != c {
					continue
				}
				pi := wsets[v].pitem.prod
				ipi := wsets[v].pitem.off + 1

				wsets[v].flag = 0
				if nolook != 0 {
					continue
				}

				ch := pi[ipi]
				ipi++
				for ch > 0 {
					// terminal symbol
					if ch < NTBASE {
						clset.set(ch)
						break
					}

					// nonterminal symbol
					clset.union(pfirst[ch-NTBASE])
					if pempty[ch-NTBASE] == 0 {
						break
					}
					ch = pi[ipi]
					ipi++
				}
				if ch <= 0 {
					clset.union(wsets[v].ws)
				}
			}

			//
			// now loop over productions derived from c
			//
			curres := pres[c-NTBASE]
			n := len(curres)

		nexts:
			// initially fill the sets
			for s := 0; s < n; s++ {
				prd := curres[s]

				//
				// put these items into the closure
				// is the item there
				//
				for v := 0; v < cwp; v++ {
					// yes, it is there
					if wsets[v].pitem.off == 0 &&
						aryeq(wsets[v].pitem.prod, prd) != 0 {
						if nolook == 0 &&
							wsets[v].ws.union(clset) {
							wsets[v].flag = 1
							work = 1
						}
						continue nexts
					}
				}

				//  not there; make a new entry
				if cwp >= len(wsets) {
					awsets := make([]Wset, cwp+WSETINC)
					copy(awsets, wsets)
					wsets = awsets
				}
				wsets[cwp].pitem = Pitem{prd, 0, prd[0], -prd[len(prd)-1]}
				wsets[cwp].flag = 1
				wsets[cwp].ws = newLkset()
				if nolook == 0 {
					work = 1
					copy(wsets[cwp].ws, clset)
				}
				cwp++
			}
		}
	}

	// have computed closure; flags are reset; return
	if cldebug != 0 && foutput != nil {
		fmt.Fprintf(foutput, "\nState %v, nolook = %v\n", i, nolook)
		for u := 0; u < cwp; u++ {
			if wsets[u].flag != 0 {
				fmt.Fprintf(foutput, "flag set\n")
			}
			wsets[u].flag = 0
			fmt.Fprintf(foutput, "\t%v", writem(wsets[u].pitem))
			fmt.Fprintln(foutput, wsets[u].ws)
			fmt.Fprintf(foutput, "\n")
		}
	}
}

// sorts last state,and sees if it equals earlier ones. returns state number
func state(c int) int {
	zzstate++
	p1 := pstate[nstate]
	p2 := pstate[nstate+1]
	if p1 == p2 {
		return 0 // null state
	}

	// sort the items
	var k, l int
	for k = p1 + 1; k < p2; k++ { // make k the biggest
		for l = k; l > p1; l-- {
			if statemem[l].pitem.prodno < statemem[l-1].pitem.prodno ||
				statemem[l].pitem.prodno == statemem[l-1].pitem.prodno &&
					statemem[l].pitem.off < statemem[l-1].pitem.off {
				s := statemem[l]
				statemem[l] = statemem[l-1]
				statemem[l-1] = s
			} else {
				break
			}
		}
	}

	size1 := p2 - p1 // size of state

	var i int
	if c >= NTBASE {
		i = ntstates[c-NTBASE]
	} else {
		i = tstates[c]
	}

look:
	for ; i != 0; i = mstates[i] {
		// get ith state
		q1 := pstate[i]
		q2 := pstate[i+1]
		size2 := q2 - q1
		if size1 != size2 {
			continue
		}
		k = p1
		for l = q1; l < q2; l++ {
			if aryeq(statemem[l].pitem.prod, statemem[k].pitem.prod) == 0 ||
				statemem[l].pitem.off != statemem[k].pitem.off {
				continue look
			}
			k++
		}

		// found it
		pstate[nstate+1] = pstate[nstate] // delete last state

		// fix up lookaheads
		if nolook != 0 {
			return i
		}
		k = p1
		for l = q1; l < q2; l++ {
			if statemem[l].look.union(statemem[k].look) {
				tystate[i] = MUSTDO
			}
			k++
		}
		return i
	}

	// state is new
	zznewstate++
	if nolook != 0 {
		errorf("yacc state/nolook error")
	}
	pstate[nstate+2] = p2
	if nstate+1 >= NSTATES {
		errorf("too many states")
	}
	if c >= NTBASE {
		mstates[nstate] = ntstates[c-NTBASE]
		ntstates[c-NTBASE] = nstate
	} else {
		mstates[nstate] = tstates[c]
		tstates[c] = nstate
	}
	tystate[nstate] = MUSTDO
	nstate++
	return nstate - 1
}

// pack state i from temp1 into amem
func apack(p []int, n int) int {
	//
	// we don't need to worry about checking because
	// we will only look at entries known to be there...
	// eliminate leading and trailing 0's
	//
	off := 0
	pp := 0
	for ; pp <= n && p[pp] == 0; pp++ {
		off--
	}

	// no actions
	if pp > n {
		return 0
	}
	for ; n > pp && p[n] == 0; n-- {
	}
	p = p[pp : n+1]

	// now, find a place for the elements from p to q, inclusive
	r := len(amem) - len(p)

nextk:
	for rr := 0; rr <= r; rr++ {
		qq := rr
		for pp = 0; pp < len(p); pp++ {
			if p[pp] != 0 {
				if p[pp] != amem[qq] && amem[qq] != 0 {
					continue nextk
				}
			}
			qq++
		}

		// we have found an acceptable k
		if pkdebug != 0 && foutput != nil {
			fmt.Fprintf(foutput, "off = %v, k = %v\n", off+rr, rr)
		}
		qq = rr
		for pp = 0; pp < len(p); pp++ {
			if p[pp] != 0 {
				if qq > memp {
					memp = qq
				}
				amem[qq] = p[pp]
			}
			qq++
		}
		if pkdebug != 0 && foutput != nil {
			for pp = 0; pp <= memp; pp += 10 {
				fmt.Fprintf(foutput, "\n")
				for qq = pp; qq <= pp+9; qq++ {
					fmt.Fprintf(foutput, "%v ", amem[qq])
				}
				fmt.Fprintf(foutput, "\n")
			}
		}
		return off + rr
	}
	errorf("no space in action table")
	return 0
}
