package myyacc

import (
	"fmt"
	"strings"
)

var lastred int                   // number of last reduction of a state
var defact = make([]int, NSTATES) // default actions of states

// no, left, right, binary assoc.
const (
	NOASC = iota
	LASC
	RASC
	BASC
)

var g2debug = 0 // debugging for go2gen

// writes state i
func wrstate(i int) {
	var j0, j1, u int
	var pp, qq int

	if len(errors) > 0 {
		actions := append([]int(nil), temp1...)
		defaultAction := ERRCODE
		if lastred != 0 {
			defaultAction = -lastred
		}
		stateTable[i] = Row{actions, defaultAction}
	}

	if foutput == nil {
		return
	}
	fmt.Fprintf(foutput, "\nstate %v\n", i)
	qq = pstate[i+1]
	for pp = pstate[i]; pp < qq; pp++ {
		fmt.Fprintf(foutput, "\t%v\n", writem(statemem[pp].pitem))
	}
	if tystate[i] == MUSTLOOKAHEAD {
		// print out empty productions in closure
		for u = pstate[i+1] - pstate[i]; u < cwp; u++ {
			if wsets[u].pitem.first < 0 {
				fmt.Fprintf(foutput, "\t%v\n", writem(wsets[u].pitem))
			}
		}
	}

	// check for state equal to another
	for j0 = 0; j0 <= nterm; j0++ {
		j1 = temp1[j0]
		if j1 != 0 {
			fmt.Fprintf(foutput, "\n\t%v  ", symName(j0))

			// shift, error, or accept
			if j1 > 0 {
				if j1 == ACCEPTCODE {
					fmt.Fprintf(foutput, "accept")
				} else if j1 == ERRCODE {
					fmt.Fprintf(foutput, "error")
				} else {
					fmt.Fprintf(foutput, "shift %v", j1)
				}
			} else {
				fmt.Fprintf(foutput, "reduce %v (src line %v)", -j1, rlines[-j1])
			}
		}
	}

	// output the final production
	if lastred != 0 {
		fmt.Fprintf(foutput, "\n\t.  reduce %v (src line %v)\n\n",
			lastred, rlines[lastred])
	} else {
		fmt.Fprintf(foutput, "\n\t.  error\n\n")
	}

	// now, output nonterminal actions
	j1 = nterm
	for j0 = 1; j0 <= nnonterm; j0++ {
		j1++
		if temp1[j1] != 0 {
			fmt.Fprintf(foutput, "\t%v  goto %v\n", symName(j0+NTBASE), temp1[j1])
		}
	}
}

// output state i
// temp1 has the actions, lastred the default
func addActions(act []int, i int) []int {
	var p, p1 int

	// find the best choice for lastred
	lastred = 0
	ntimes := 0
	for j := 0; j <= nterm; j++ {
		if temp1[j] >= 0 {
			continue
		}
		if temp1[j]+lastred == 0 {
			continue
		}
		// count the number of appearances of temp1[j]
		count := 0
		tred := -temp1[j]
		levprd[tred] |= REDFLAG
		for p = 0; p <= nterm; p++ {
			if temp1[p]+tred == 0 {
				count++
			}
		}
		if count > ntimes {
			lastred = tred
			ntimes = count
		}
	}

	//
	// for error recovery, arrange that, if there is a shift on the
	// error recovery token, `error', that the default be the error action
	//
	if temp1[2] > 0 {
		lastred = 0
	}

	// clear out entries in temp1 which equal lastred
	// count entries in optst table
	n := 0
	for p = 0; p <= nterm; p++ {
		p1 = temp1[p]
		if p1+lastred == 0 {
			temp1[p] = 0
			p1 = 0
		}
		if p1 > 0 && p1 != ACCEPTCODE && p1 != ERRCODE {
			n++
		}
	}

	wrstate(i)
	defact[i] = lastred
	flag := 0
	os := make([]int, n*2)
	n = 0
	for p = 0; p <= nterm; p++ {
		p1 = temp1[p]
		if p1 != 0 {
			if p1 < 0 {
				p1 = -p1
			} else if p1 == ACCEPTCODE {
				p1 = -1
			} else if p1 == ERRCODE {
				p1 = 0
			} else {
				os[n] = p
				n++
				os[n] = p1
				n++
				zzacent++
				continue
			}
			if flag == 0 {
				act = append(act, -1, i)
			}
			flag++
			act = append(act, p, p1)
			zzexcp++
		}
	}
	if flag != 0 {
		defact[i] = -2
		act = append(act, -2, lastred)
	}
	optst[i] = os
	return act
}

// print the output for the states
func output() {
	var c, u, v int

	if !lflag {
		fmt.Fprintf(ftable, "\n//line yacctab:1")
	}
	var actions []int

	if len(errors) > 0 {
		stateTable = make([]Row, nstate)
	}

	noset := newLkset()

	// output the stuff for state i
	for i := 0; i < nstate; i++ {
		nolook = 0
		if tystate[i] != MUSTLOOKAHEAD {
			nolook = 1
		}
		closure(i)

		// output actions
		nolook = 1
		aryfil(temp1, nterm+nnonterm+1, 0)
		for u = 0; u < cwp; u++ {
			c = wsets[u].pitem.first
			if c > 1 && c < NTBASE && temp1[c] == 0 {
				for v = u; v < cwp; v++ {
					if c == wsets[v].pitem.first {
						putitem(wsets[v].pitem, noset)
					}
				}
				temp1[c] = state(c)
			} else if c > NTBASE {
				c -= NTBASE
				if temp1[c+nterm] == 0 {
					temp1[c+nterm] = amem[indgo[i]+c]
				}
			}
		}
		if i == 1 {
			temp1[1] = ACCEPTCODE
		}

		// now, we have the shifts; look at the reductions
		lastred = 0
		for u = 0; u < cwp; u++ {
			c = wsets[u].pitem.first
			// reduction
			if c > 0 {
				continue
			}
			lastred = -c
			us := wsets[u].ws
			for k := 0; k <= nterm; k++ {
				if !us.check(k) {
					continue
				}
				if temp1[k] == 0 {
					temp1[k] = c
				} else if temp1[k] < 0 { // reduce/reduce conflict
					if foutput != nil {
						fmt.Fprintf(foutput,
							"\n %v: reduce/reduce conflict  (red'ns "+
								"%v and %v) on %v",
							i, -temp1[k], lastred, symName(k))
					}
					if -temp1[k] > lastred {
						temp1[k] = -lastred
					}
					zzrrconf++
				} else {
					// potential shift/reduce conflict
					precftn(lastred, k, i)
				}
			}
		}
		actions = addActions(actions, i)
	}

	arrayOutColumns("Exca", actions, 2, false)
	fmt.Fprintf(ftable, "\n")
	ftable.WriteRune('\n')
	fmt.Fprintf(ftable, "const %sPrivate = %v\n", prefix, PRIVATE)
}

// decide a shift/reduce conflict by precedence.
// r is a rule number, t a token number
// the conflict is in state s
// temp1[t] is changed to reflect the action
func precftn(r, t, s int) {
	action := NOASC

	lp := levprd[r]
	lt := preds[t]
	if PLEVEL(lt) == 0 || PLEVEL(lp) == 0 {
		// conflict
		if foutput != nil {
			fmt.Fprintf(foutput,
				"\n%v: shift/reduce conflict (shift %v(%v), red'n %v(%v)) on %v",
				s, temp1[t], PLEVEL(lt), r, PLEVEL(lp), symName(t))
		}
		zzsrconf++
		return
	}
	if PLEVEL(lt) == PLEVEL(lp) {
		action = ASSOC(lt)
	} else if PLEVEL(lt) > PLEVEL(lp) {
		action = RASC // shift
	} else {
		action = LASC
	} // reduce
	switch action {
	case BASC: // error action
		temp1[t] = ERRCODE
	case LASC: // reduce
		temp1[t] = -r
	}
}

// output the gotos for the nontermninals
func go2out() {
	for i := 1; i <= nnonterm; i++ {
		go2gen(i)

		// find the best one to make default
		best := -1
		times := 0

		// is j the most frequent
		for j := 0; j < nstate; j++ {
			if tystate[j] == 0 {
				continue
			}
			if tystate[j] == best {
				continue
			}

			// is tystate[j] the most frequent
			count := 0
			cbest := tystate[j]
			for k := j; k < nstate; k++ {
				if tystate[k] == cbest {
					count++
				}
			}
			if count > times {
				best = cbest
				times = count
			}
		}

		// best is now the default entry
		zzgobest += times - 1
		n := 0
		for j := 0; j < nstate; j++ {
			if tystate[j] != 0 && tystate[j] != best {
				n++
			}
		}
		goent := make([]int, 2*n+1)
		n = 0
		for j := 0; j < nstate; j++ {
			if tystate[j] != 0 && tystate[j] != best {
				goent[n] = j
				n++
				goent[n] = tystate[j]
				n++
				zzgoent++
			}
		}

		// now, the default
		if best == -1 {
			best = 0
		}

		zzgoent++
		goent[n] = best
		yypgo[i] = goent
	}
}

// output the gotos for nonterminal c
func go2gen(c int) {
	var i, cc, p, q int

	// first, find nonterminals with gotos on c
	aryfil(temp1, nnonterm+1, 0)
	temp1[c] = 1
	work := 1
	for work != 0 {
		work = 0
		for i = 0; i < nprod; i++ {
			// cc is a nonterminal with a goto on c
			cc = prdptr[i][1] - NTBASE
			if cc >= 0 && temp1[cc] != 0 {
				// thus, the left side of production i does too
				cc = prdptr[i][0] - NTBASE
				if temp1[cc] == 0 {
					work = 1
					temp1[cc] = 1
				}
			}
		}
	}

	// now, we have temp1[c] = 1 if a goto on c in closure of cc
	if g2debug != 0 && foutput != nil {
		fmt.Fprintf(foutput, "%v: gotos on ", nonterms[c].name)
		for i = 0; i <= nnonterm; i++ {
			if temp1[i] != 0 {
				fmt.Fprintf(foutput, "%v ", nonterms[i].name)
			}
		}
		fmt.Fprintf(foutput, "\n")
	}

	// now, go through and put gotos into tystate
	aryfil(tystate, nstate, 0)
	for i = 0; i < nstate; i++ {
		q = pstate[i+1]
		for p = pstate[i]; p < q; p++ {
			cc = statemem[p].pitem.first
			if cc >= NTBASE {
				// goto on c is possible
				if temp1[cc-NTBASE] != 0 {
					tystate[i] = amem[indgo[i]+c]
					break
				}
			}
		}
	}
}

func arrayOutColumns(s string, v []int, columns int, allowUnsigned bool) {
	s = prefix + s
	ftable.WriteRune('\n')
	minType := minType(v, allowUnsigned)
	fmt.Fprintf(ftable, "var %v = [...]%s{", s, minType)
	for i, val := range v {
		if i%columns == 0 {
			fmt.Fprintf(ftable, "\n\t")
		} else {
			ftable.WriteRune(' ')
		}
		fmt.Fprintf(ftable, "%d,", val)
	}
	fmt.Fprintf(ftable, "\n}\n")
}

func arout(s string, v []int, n int) {
	arrayOutColumns(s, v[:n], 10, true)
}

// output the summary on y.output
func summary() {
	if foutput != nil {
		fmt.Fprintf(foutput, "\n%v terminals, %v nonterminals\n", nterm, nnonterm+1)
		fmt.Fprintf(foutput, "%v grammar rules, %v/%v states\n", nprod, nstate, NSTATES)
		fmt.Fprintf(foutput, "%v shift/reduce, %v reduce/reduce conflicts reported\n", zzsrconf, zzrrconf)
		fmt.Fprintf(foutput, "%v working sets used\n", len(wsets))
		fmt.Fprintf(foutput, "memory: parser %v/%v\n", memp, ACTSIZE)
		fmt.Fprintf(foutput, "%v extra closures\n", zzclose-2*nstate)
		fmt.Fprintf(foutput, "%v shift entries, %v exceptions\n", zzacent, zzexcp)
		fmt.Fprintf(foutput, "%v goto entries\n", zzgoent)
		fmt.Fprintf(foutput, "%v entries saved by goto default\n", zzgobest)
	}
	if zzsrconf != 0 || zzrrconf != 0 {
		fmt.Printf("\nconflicts: ")
		if zzsrconf != 0 {
			fmt.Printf("%v shift/reduce", zzsrconf)
		}
		if zzsrconf != 0 && zzrrconf != 0 {
			fmt.Printf(", ")
		}
		if zzrrconf != 0 {
			fmt.Printf("%v reduce/reduce", zzrrconf)
		}
		fmt.Printf("\n")
	}
}

// write optimizer summary
func osummary() {
	if foutput == nil {
		return
	}
	i := 0
	for p := maxa; p >= 0; p-- {
		if amem[p] == 0 {
			i++
		}
	}

	fmt.Fprintf(foutput, "Optimizer space used: output %v/%v\n", maxa+1, ACTSIZE)
	fmt.Fprintf(foutput, "%v table entries, %v zero\n", maxa+1, i)
	fmt.Fprintf(foutput, "maximum spread: %v, maximum offset: %v\n", maxspr, maxoff)
}

func usage() {
	fmt.Fprintf(stderr, "usage: yacc [-o output] [-v parsetable] input\n")
	exit(1)
}

// this version is for limbo
// write out the optimized parser
func aoutput() {
	ftable.WriteRune('\n')
	fmt.Fprintf(ftable, "const %sLast = %v\n", prefix, maxa+1)
	arout("Act", amem, maxa+1)
	arout("Pact", indgo, nstate)
	arout("Pgo", pgo, nnonterm+1)
}

// put out other arrays, copy the parsers
func others() {
	var i, j int

	arout("R1", levprd, nprod)
	aryfil(temp1, nprod, 0)

	//
	//yyr2 is the number of rules for each production
	//
	for i = 1; i < nprod; i++ {
		temp1[i] = len(prdptr[i]) - 2
	}
	arout("R2", temp1, nprod)

	aryfil(temp1, nstate, -1000)
	for i = 0; i <= nnonterm; i++ {
		for j := tstates[i]; j != 0; j = mstates[j] {
			temp1[j] = i
		}
	}
	for i = 0; i <= nnonterm; i++ {
		for j = ntstates[i]; j != 0; j = mstates[j] {
			temp1[j] = -i
		}
	}
	arout("Chk", temp1, nstate)
	arrayOutColumns("Def", defact[:nstate], 10, false)

	// put out token translation tables
	// table 1 has 0-256
	aryfil(temp1, 256, 0)
	c := 0
	for i = 1; i <= nterm; i++ {
		j = terms[i].value
		if j >= 0 && j < 256 {
			if temp1[j] != 0 {
				fmt.Print("yacc bug -- cannot have 2 different Ts with same value\n")
				fmt.Printf("	%s and %s\n", terms[i].name, terms[temp1[j]].name)
				nerrors++
			}
			temp1[j] = i
			if j > c {
				c = j
			}
		}
	}
	for i = 0; i <= c; i++ {
		if temp1[i] == 0 {
			temp1[i] = YYLEXUNK
		}
	}
	arout("Tok1", temp1, c+1)

	// table 2 has PRIVATE-PRIVATE+256
	aryfil(temp1, 256, 0)
	c = 0
	for i = 1; i <= nterm; i++ {
		j = terms[i].value - PRIVATE
		if j >= 0 && j < 256 {
			if temp1[j] != 0 {
				fmt.Print("yacc bug -- cannot have 2 different Ts with same value\n")
				fmt.Printf("	%s and %s\n", terms[i].name, terms[temp1[j]].name)
				nerrors++
			}
			temp1[j] = i
			if j > c {
				c = j
			}
		}
	}
	arout("Tok2", temp1, c+1)

	// table 3 has everything else
	ftable.WriteRune('\n')
	var v []int
	for i = 1; i <= nterm; i++ {
		j = terms[i].value
		if j >= 0 && j < 256 {
			continue
		}
		if j >= PRIVATE && j < 256+PRIVATE {
			continue
		}

		v = append(v, j, i)
	}
	v = append(v, 0)
	arout("Tok3", v, len(v))
	fmt.Fprintf(ftable, "\n")

	// Custom error messages.
	fmt.Fprintf(ftable, "\n")
	fmt.Fprintf(ftable, "var %sErrorMessages = [...]struct {\n", prefix)
	fmt.Fprintf(ftable, "\tstate int\n")
	fmt.Fprintf(ftable, "\ttoken int\n")
	fmt.Fprintf(ftable, "\tmsg   string\n")
	fmt.Fprintf(ftable, "}{\n")
	for _, error := range errors {
		lineno = error.lineno
		state, token := runMachine(error.tokens)
		fmt.Fprintf(ftable, "\t{%v, %v, %s},\n", state, token, error.msg)
	}
	fmt.Fprintf(ftable, "}\n")

	// copy parser text
	ch := getrune(finput)
	for ch != EOF {
		ftable.WriteRune(ch)
		ch = getrune(finput)
	}

	// copy yaccpar
	if !lflag {
		fmt.Fprintf(ftable, "\n//line yaccpar:1\n")
	}

	parts := strings.SplitN(yaccpar, prefix+"run()", 2)
	fmt.Fprintf(ftable, "%v", parts[0])
	ftable.Write(fcode.Bytes())
	fmt.Fprintf(ftable, "%v", parts[1])
}

func runMachine(tokens []string) (state, token int) {
	var stack []int
	i := 0
	token = -1

Loop:
	if token < 0 {
		token = chfind(2, tokens[i])
		i++
	}

	row := stateTable[state]

	c := token
	if token >= NTBASE {
		c = token - NTBASE + nterm
	}
	action := row.actions[c]
	if action == 0 {
		action = row.defaultAction
	}

	switch {
	case action == ACCEPTCODE:
		errorf("tokens are accepted")
		return
	case action == ERRCODE:
		if token >= NTBASE {
			errorf("error at non-terminal token %s", symName(token))
		}
		return
	case action > 0:
		// Shift to state action.
		stack = append(stack, state)
		state = action
		token = -1
		goto Loop
	default:
		// Reduce by production -action.
		prod := prdptr[-action]
		if rhsLen := len(prod) - 2; rhsLen > 0 {
			n := len(stack) - rhsLen
			state = stack[n]
			stack = stack[:n]
		}
		if token >= 0 {
			i--
		}
		token = prod[0]
		goto Loop
	}
}
