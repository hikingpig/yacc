package myacc2

var gotoIdx = make([]int, NSTATES)     // idx to compressed goto rows. each row is for a newly closed state
var actionStore = make([]int, ACTSIZE) // store the packed shifts and gotos
var packedGotos [][]int                // the goto table compressed vertically
var pgoIdx []int                       // idx of compressed goto-columns (packedGotos) stored in actionStore for each non-terminal symbol
var lastActIdx int                     // last idx of a shift or goto in actionStore
var maxoff = 0                         // shiftoff or max state that has a goto
var shiftoff int                       // a shift-row for a state has a minimum term. shiftoff is the max of those minimum terms for all rows
var maxspr = 0                         // maximum spread, the maximum of terminal symbol that induces a shift
var shiftIdx []int                     // idx of compressed shift-rows (shifts) stored in actionStore for each state
var yyFlag = -1000                     // mark invalid shift

// compress newly derived goto row after closing state i. each column is for a non-terminal symbol
func packGotosRow() int {
	m := 1 // exclude $accept
	n := nontermN
	// skip leading zeros
	for m < n && trans[m] == 0 {
		m++
	}
	if m >= n {
		return 0
	}
	// skip trailing zeros
	for n > m && trans[n] == 0 {
		n--
	}
	trans = trans[m : n+1]
	maxStart := len(actionStore) - len(trans)
nextStart:
	for start := 0; start < maxStart; start++ {
		// find spot with enough empty room for new goTos
		for i, j := 0, start; i < len(trans); i, j = i+1, j+1 {
			if trans[i] != 0 && actionStore[j] != 0 && trans[i] != actionStore[j] {
				continue nextStart
			}
		}
		// found it, store action
		for i, j := 0, start; i < len(trans); i, j = i+1, j+1 {
			if trans[i] != 0 {
				actionStore[j] = trans[i]
				if j > lastActIdx {
					lastActIdx = j
				}
			}
		}

		return -m + start
	}
	// can not find empty spots for new goTos
	return 0
}

// compress goto table by column. each column is for a non-terminal symbol.
func packGotosByCol() {
	nextStates := make([]int, nstate)
	for i := 1; i <= nontermN; i++ {
		packCol(i, nextStates)
		// compute the default goto
		defGoto := -1
		maxCount := 0
		goToCounts := map[int]int{}
		for _, nextState := range nextStates {
			goToCounts[nextState]++
			if goToCounts[nextState] > maxCount {
				maxCount = goToCounts[nextState]
				defGoto = nextState
			}
		}
		n := 0
		for _, s := range nextStates {
			if s != 0 && s != defGoto {
				n++ // count the goto that are not default
			}
		}
		col := make([]int, 2*n+1)
		n = 0
		for currentstate, nextstate := range nextStates {
			if nextstate != 0 && nextstate != defGoto {
				col[n] = currentstate
				n++
				col[n] = nextstate
				n++
			}
		}
		if defGoto == -1 {
			defGoto = 0
		}
		col[n] = defGoto
		packedGotos[i] = col
	}
}

// pack a column of goto table for non-terminal symbol s
// find all the pairs (current state, next state) transitions induced by s
func packCol(s int, nextStates []int) {
	deriveS := make([]bool, nontermN+1) // track if a symbol derives s during closure
	deriveS[s] = true                   // s derives itself
	changed := true
	var a int
	for changed {
		changed = false
		for _, p := range allPrds {
			a = p.prd[1] - NTBASE
			if a >= 0 && deriveS[a] {
				a = p.prd[0] - NTBASE
				if !deriveS[a] {
					deriveS[a] = true
					changed = true
				}
			}
		}
	}
	fill(nextStates, nstate, 0)
	// check if first symbol in each kernel item derive s
	// if so, retrieve the goto from actionStore and move on
	for i := range nextStates {
		for _, itemI := range kernls[kernlp[i]:kernlp[i+1]] {
			first := itemI.first
			if first >= NTBASE && deriveS[first] {
				nextStates[i] = actionStore[gotoIdx[i]+s]
				break
			}
		}
	}
}

// store a column of packed goto table for a non-terminal i into actionStore
func storeGotoCol(i int) {
	col := packedGotos[i]
	last := len(col) - 1
	var idx int
nextStart:
	// maxoff is not used because the default goto will always be stored
	// at beginning of available space
	for start, a := range actionStore {
		if a != 0 {
			continue
		}
		for r := 0; r < last; r += 2 {
			idx = start + col[r] + 1
			if idx > lastActIdx {
				lastActIdx = idx
			}
			if lastActIdx >= ACTSIZE {
				errorf("actionStore overflow")
			}
			if actionStore[idx] != 0 {
				continue nextStart
			}
		}
		// find spot, store goto column into actionStore
		actionStore[start] = col[last] // default goto. how can we check it is correct for other state?
		if start > lastActIdx {        // edge-case: last == 0
			lastActIdx = a
		}
		for r := 0; r < last; r += 2 {
			s := start + col[r] + 1
			actionStore[s] = col[r+1]
		}
		pgoIdx[i] = start
	}
}

// store the shifts in a row in action table for a state i into actionStore
func storeShifts(i int) {
	row := shifts[i]
	l := len(row)
	var idx int
nextStart:
	// shiftoff can help save a few leading 0s when min(row[r]) > 0
	// but it is only effective for maximum one time!
	for start := -shiftoff; start < ACTSIZE; start++ {
		mayConflict := false // mark conflict may arise if another state has taken the same spot
		for r := 0; r < l; r += 2 {
			idx = row[r] + start
			if idx < 0 {
				continue nextStart
			}
			if idx >= ACTSIZE {
				errorf("actionStore overflow")
			}
			if actionStore[idx] == 0 && row[r+1] != 0 {
				mayConflict = true
			} else if actionStore[idx] != row[r+1] {
				continue nextStart
			}
		}
		// check if any previously processed state has identical shifts
		for j := 0; j < nstate; j++ {
			if shiftIdx[j] == start {
				if mayConflict {
					continue nextStart
				}
				if l == len(shifts[j]) { // identical shifts. assign idx. done
					shiftIdx[i] = start
					return
				}
				continue nextStart
			}
		}
		// find the spot, store shifts into actionStore
		for r := 0; r < l; r += 2 {
			idx := row[r] + start
			if idx > lastActIdx {
				lastActIdx = idx
			}
			if actionStore[idx] != 0 && actionStore[idx] != row[r+1] {
				errorf("clobber of actionStore, pos'n %d, by %d", idx, row[r+1])
			}
			actionStore[idx] = row[r+1]
		}
		shiftIdx[i] = start
		return
	}
	errorf("Error; failure to place a state %v", i)
}

// store the shift-row for each state or goto-column for each non-terminal symbol into actionStore
// the one with highest priority will be processed first. the one with 0 priority will be skipped
func storeShiftsAndGotos() {
	var i int
	var row []int
	var maxTerm, minTerm int                // maximum and minimum terminals that induce shifts for a state
	var shiftPriority = make([]int, nstate) // shift row with highest priority will be processed first
	// compute shiftPriority
	for i = 0; i < nstate; i++ {
		minTerm = 32000
		maxTerm = 0
		row = shifts[i]
		for j := 0; j < len(row); j += 2 {
			if row[j] > maxTerm {
				maxTerm = row[j]
			}
			if row[j] < minTerm {
				minTerm = row[j]
			}
		}
		// if row is empty, minTerm > maxTerm!
		if minTerm <= maxTerm {
			if minTerm > maxoff {
				maxoff = minTerm
			}
		}
		shiftPriority[i] = len(row) + 2*maxTerm // can be 0. can be skipped
		if maxTerm > maxspr {
			maxspr = maxTerm
		}
	}
	shiftoff = maxoff
	var gotoPriority = make([]int, nontermN+1) // goto column with highest priority will be processed first
	var maxState int                           // the highest state that has a goto for a non-terminal symbol
	var col []int
	// compute gotoPriority
	for i = 0; i <= nontermN; i++ {
		gotoPriority[i] = 1 // can NOT be 0. can not be skipped
		maxState = 0
		col = packedGotos[i]
		for j := 0; j < len(col)-1; j += 2 {
			gotoPriority[i] += 2
			if col[j] > maxState {
				maxState = col[j]
			}
		}
		gotoPriority[i] += 2 * maxState
		if maxState > maxoff {
			maxoff = maxState
		}
	}
	// clear the action store
	for i = 0; i < ACTSIZE; i++ {
		actionStore[i] = 0
	}
	lastActIdx = 0
	// prepare shiftIdx
	for i = 0; i < nstate; i++ {
		shiftIdx[i] = yyFlag // no valid shift initially
	}
	// prepare packed gotoIdx
	pgoIdx = make([]int, nontermN+1)
	pgoIdx[0] = 0 // $accept induces no goto!
	// choose shift-row or goto-col with highest priority to process
	i = nxti(gotoPriority, shiftPriority)
	for i != NOMORE {
		if i >= 0 {
			storeShifts(i)
		} else {
			storeGotoCol(i)
		}
		i = nxti(gotoPriority, shiftPriority)
	}
}

// choose i with highest priority to be processed
// return negative value if gotoPriority wins
// return non-negative value if shiftPriority wins
func nxti(gotoPriority, shiftPriority []int) int {
	maxPriority := 0
	maxi := 0
	for i := 1; i <= nontermN; i++ {
		if gotoPriority[i] >= maxPriority {
			maxPriority = gotoPriority[i]
			maxi = -i
		}
	}
	for i := 0; i < nstate; i++ {
		if shiftPriority[i] >= maxPriority {
			maxPriority = shiftPriority[i]
			maxi = i
		}
	}
	if maxPriority == 0 {
		return NOMORE
	}
	// clear the maxPriority
	if maxi >= 0 {
		shiftPriority[maxi] = 0
	} else {
		gotoPriority[-maxi] = 0
	}
	return maxi
}
