package myyacc

import (
	"strconv"
	"unicode/utf8"
)

// Symbol is a symbol in a production
type Symbol struct {
	name    string
	value   int
	isconst bool
}

var nterm int      // number of terminal symbols
var terms []Symbol // store of terminal symbols
var preds []int    // precedences of terminal symbols

var nnonterm = -1     // number of non-terminal symbols
var nonterms []Symbol // list of non-terminal symbols

var extval = 0 // track the value of long terminal symbol (> 1 character)

/*
addTerm adds a terminal symbol to terms
*/
func addTerm(s string) int {
	nterm++
	if nterm > len(terms) {
		extend(&terms, SYMINC)
		extend(&preds, SYMINC)
	}
	terms[nterm].name = s
	terms[nterm].isconst = true
	preds[nterm] = 0
	var val int
	// literal token. accept only single character token
	if s[0] == '\'' || s[0] == '"' {
		s, err := strconv.Unquote(s)
		assert(err == nil, "invalid token: %s", err)
		assert(utf8.RuneCountInString(s) == 1, "character token too long: %s", s)
		val = int([]rune(s)[0])
		assert(val == 0, "token value 0 is illegal")
		terms[nterm].isconst = false
		// long terminal symbol
	} else {
		val = extval
		extval++
		if s[0] == '$' {
			terms[nterm].isconst = false
		}
	}
	terms[nterm].value = val
	return nterm
}

/*
addNonterm adds a non-terminal symbol to nonterms
*/
func addNonterm(s string) int {
	nnonterm++
	if nnonterm > len(nonterms) {
		extend(&nonterms, SYMINC)
	}
	nonterms[nnonterm] = Symbol{name: s} // non-terminal is not constant and has no value!
	return NTBASE + nnonterm
}

/* addInitSymbols adds the symbols required internally for yacc */
func addInitSyms() {
	addTerm("$end")  // $end symbol's value is 0
	extval = PRIVATE // all later terminal symbol's value are from PRIVATE
	addTerm("error")
	addNonterm("$accept")
	addTerm("$unk")
}

/* symType gets the type of a symbol */
// ====== VERY OBSCURE. replace it
func symType(t int) int {
	var v int
	var s string

	if t >= NTBASE { // nonterminal
		v = nonterms[t-NTBASE].value // ========== TODO: consider not using numbers for symbols? use symbols directly? because the parse tables doesn't required symbols to be numbers!!!
		s = nonterms[t-NTBASE].name
	} else { // terminal
		v = TYPE(preds[t]) // ================= TODO: inconsistent use of fields. should do the same separate them!
		s = terms[t].name
	}
	if v <= 0 {
		errorf("must specify type for %v", s)
	}
	return v
}

// return the name of symbol i
func symName(i int) string {
	var s string

	if i >= NTBASE {
		s = nonterms[i-NTBASE].name
	} else {
		s = terms[i].name
	}
	return s
}

/*
findTerm find a terminal in the terms.
if not found, return -1
*/
func findTerm(s string) int {
	for i := 0; i <= nterm; i++ {
		if s == terms[i].name {
			return i
		}
	}
	return -1
}

/*
findNonTerm find a non-terminal in the terms.
if not found, return -1
*/

func findNonTerm(s string) int {
	for i := 0; i <= nnonterm; i++ {
		if s == nonterms[i].name {
			return NTBASE + i
		}
	}
	return -1
}

// this functions is very obscure. try to split it up!
func chfind(t int, s string) int {
	if s[0] == '"' || s[0] == '\'' { // this will always terminal token
		t = 0
	}
	for i := 0; i <= nterm; i++ {
		if s == terms[i].name {
			return i
		}
	}
	for i := 0; i <= nnonterm; i++ {
		if s == nonterms[i].name {
			return NTBASE + i
		}
	}

	// cannot find name
	if t > 1 {
		errorf("%v should have been defined earlier", s)
	}
	if t > 0 {
		return addNonterm(s)
	}
	return addTerm(s)
}
