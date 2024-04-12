package myacc2

var gotoIdx = make([]int, NSTATES)
var actionStore = make([]int, ACTSIZE)
var lastGoto int // last goto's index in goTos
var packedGotos [][]int
var pgoIdx []int // idx of compressed gotos stored in actionStore for a non-terminal symbol

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
				if j > lastGoto {
					lastGoto = j
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
