package myacc2

var gotoIdx = make([]int, NSTATES)
var actionStore = make([]int, ACTSIZE)
var packedGotos [][]int
var pgoIdx []int // idx of compressed gotos stored in actionStore for a non-terminal symbol
var lastActIdx int
var maxoff int     // ===== comment later
var shiftIdx []int // ======= should be intialized somewhere!!!!
// compress goto table by row, each row is for a state. each column is for a non-terminal symbol
func packGotosRow(trans []int, n int) int {
	m := 0
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
// each item store is a pair of (current state, next state)
// the last value is default goto
func packCol(s int, nextStates []int) {
	derived := make([]bool, nontermN+1) // track if a symbol derives s during closure
	derived[s] = true                   // s derives itself
	changed := true
	var a int
	for changed {
		changed = false
		for _, p := range allPrds {
			a = p.prd[1] - NTBASE
			if a >= 0 && derived[a] {
				a = p.prd[0] - NTBASE
				if !derived[a] {
					derived[a] = true
					changed = true
				}
			}
		}
	}
	fill(nextStates, nstate, 0)
	for i := range nextStates {
		for _, itemI := range kernls[kernlp[i]:kernlp[i+1]] {
			first := itemI.first
			if first >= NTBASE && derived[first] { // if first derives s
				nextStates[i] = actionStore[gotoIdx[i]+s] // s induced a goto on this state
				break                                     // move to next state
			}
		}
	}
}

// store a column of packed goto table for a non-terminal i into actionStore
func storeGotoCol(i int) {
	// ============= set ggreed somewhere else
	col := packedGotos[i]
	last := len(col) - 1
	var idx int
nextStart:
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
	// ====== clear tystate somewhere else
	row := shifts[i]
	l := len(row)
	var idx int
nextStart:
	for start := -maxoff; start < ACTSIZE; start++ { // ====== roll of maxoff clarified later!
		mayConflict := false
		for r := 0; r < l; r += 2 {
			idx = row[r] + start
			if idx < 0 || idx > ACTSIZE { // ============= shouldn't it error here???
				continue nextStart
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
			// ========== will this happen? we check everything already!!!
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
