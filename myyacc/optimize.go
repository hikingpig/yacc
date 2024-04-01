package myyacc

import "fmt"

// optimizer arrays

var yypgo [][]int
var optst [][]int
var ggreed []int
var pgo []int
var adb = 0 // debugging for callopt
// output parser flags
const yyFlag = -1000

// in order to free up the mem and amem arrays for the optimizer,
// and still be able to output yyr1, etc., after the sizes of
// the action array is known, we hide the nonterminals
// derived by productions in levprd.
func hideprod() {
	nred := 0
	levprd[0] = 0
	for i := 1; i < nprod; i++ {
		if (levprd[i] & REDFLAG) == 0 {
			if foutput != nil {
				fmt.Fprintf(foutput, "Rule not reduced: %v\n",
					writem(Pitem{prdptr[i], 0, 0, i}))
			}
			fmt.Printf("rule %v never reduced\n", writem(Pitem{prdptr[i], 0, 0, i}))
			nred++
		}
		levprd[i] = prdptr[i][0] - NTBASE
	}
	if nred != 0 {
		fmt.Printf("%v rules never reduced\n", nred)
	}
}

func callopt() {
	var j, k, p, q, i int
	var v []int

	pgo = make([]int, nnonterm+1)
	pgo[0] = 0
	maxoff = 0
	maxspr = 0
	for i = 0; i < nstate; i++ {
		k = 32000
		j = 0
		v = optst[i]
		q = len(v)
		for p = 0; p < q; p += 2 {
			if v[p] > j {
				j = v[p]
			}
			if v[p] < k {
				k = v[p]
			}
		}

		// nontrivial situation
		if k <= j {
			// j is now the range
			//			j -= k;			// call scj
			if k > maxoff {
				maxoff = k
			}
		}
		tystate[i] = q + 2*j
		if j > maxspr {
			maxspr = j
		}
	}

	// initialize ggreed table
	ggreed = make([]int, nnonterm+1)
	for i = 1; i <= nnonterm; i++ {
		ggreed[i] = 1
		j = 0

		// minimum entry index is always 0
		v = yypgo[i]
		q = len(v) - 1
		for p = 0; p < q; p += 2 {
			ggreed[i] += 2
			if v[p] > j {
				j = v[p]
			}
		}
		ggreed[i] = ggreed[i] + 2*j
		if j > maxoff {
			maxoff = j
		}
	}

	// now, prepare to put the shift actions into the amem array
	for i = 0; i < ACTSIZE; i++ {
		amem[i] = 0
	}
	maxa = 0
	for i = 0; i < nstate; i++ {
		if tystate[i] == 0 && adb > 1 {
			fmt.Fprintf(ftable, "State %v: null\n", i)
		}
		indgo[i] = yyFlag
	}

	i = nxti()
	for i != NOMORE {
		if i >= 0 {
			stin(i)
		} else {
			gin(-i)
		}
		i = nxti()
	}

	// print amem array
	if adb > 2 {
		for p = 0; p <= maxa; p += 10 {
			fmt.Fprintf(ftable, "%v  ", p)
			for i = 0; i < 10; i++ {
				fmt.Fprintf(ftable, "%v  ", amem[p+i])
			}
			ftable.WriteRune('\n')
		}
	}

	aoutput()
	osummary()
}

// finds the next i
func nxti() int {
	max := 0
	maxi := 0
	for i := 1; i <= nnonterm; i++ {
		if ggreed[i] >= max {
			max = ggreed[i]
			maxi = -i
		}
	}
	for i := 0; i < nstate; i++ {
		if tystate[i] >= max {
			max = tystate[i]
			maxi = i
		}
	}
	if max == 0 {
		return NOMORE
	}
	return maxi
}

func gin(i int) {
	var s int

	// enter gotos on nonterminal i into array amem
	ggreed[i] = 0

	q := yypgo[i]
	nq := len(q) - 1

	// now, find amem place for it
nextgp:
	for p := 0; p < ACTSIZE; p++ {
		if amem[p] != 0 {
			continue
		}
		for r := 0; r < nq; r += 2 {
			s = p + q[r] + 1
			if s > maxa {
				maxa = s
				if maxa >= ACTSIZE {
					errorf("a array overflow")
				}
			}
			if amem[s] != 0 {
				continue nextgp
			}
		}

		// we have found amem spot
		amem[p] = q[nq]
		if p > maxa {
			maxa = p
		}
		for r := 0; r < nq; r += 2 {
			s = p + q[r] + 1
			amem[s] = q[r+1]
		}
		pgo[i] = p
		if adb > 1 {
			fmt.Fprintf(ftable, "Nonterminal %v, entry at %v\n", i, pgo[i])
		}
		return
	}
	errorf("cannot place goto %v\n", i)
}

func stin(i int) {
	var s int

	tystate[i] = 0

	// enter state i into the amem array
	q := optst[i]
	nq := len(q)

nextn:
	// find an acceptable place
	for n := -maxoff; n < ACTSIZE; n++ {
		flag := 0
		for r := 0; r < nq; r += 2 {
			s = q[r] + n
			if s < 0 || s > ACTSIZE {
				continue nextn
			}
			if amem[s] == 0 {
				flag++
			} else if amem[s] != q[r+1] {
				continue nextn
			}
		}

		// check the position equals another only if the states are identical
		for j := 0; j < nstate; j++ {
			if indgo[j] == n {

				// we have some disagreement
				if flag != 0 {
					continue nextn
				}
				if nq == len(optst[j]) {

					// states are equal
					indgo[i] = n
					if adb > 1 {
						fmt.Fprintf(ftable, "State %v: entry at"+
							"%v equals state %v\n",
							i, n, j)
					}
					return
				}

				// we have some disagreement
				continue nextn
			}
		}

		for r := 0; r < nq; r += 2 {
			s = q[r] + n
			if s > maxa {
				maxa = s
			}
			if amem[s] != 0 && amem[s] != q[r+1] {
				errorf("clobber of a array, pos'n %v, by %v", s, q[r+1])
			}
			amem[s] = q[r+1]
		}
		indgo[i] = n
		if adb > 1 {
			fmt.Fprintf(ftable, "State %v: entry at %v\n", i, indgo[i])
		}
		return
	}
	errorf("Error; failure to place state %v", i)
}
