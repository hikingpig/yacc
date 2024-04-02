package myacc2

import "strconv"

type SymTerm struct {
	name    string
	isConst bool
	value   int
}

type SymNonterm struct {
	name string
	typ  int // hold type value of non-term
}

var terms []SymTerm
var nonterms []SymNonterm

// store non-term into the nonterms
func defineNT(s string) int {
	nontermN++
	if nontermN >= len(nonterms) {
		extend(&nonterms, SYMINC)
	}
	nonterms[nontermN] = SymNonterm{name: s}
	return NTBASE + nontermN
}

// store terminal symbol into the terms
// value of single literal terminal is extracted from its character
// value of other terminal is assigned from extval. extval is increased after each used
func defineTerm(s string) int {
	termN++
	if termN > len(terms) {
		extend(&terms, SYMINC)
		extend(&aptTerm, SYMINC)
	}
	terms[termN].name = s
	terms[termN].isConst = true
	aptTerm[termN] = 0
	val := 0
	// single character literal term-symbol
	if s[0] == '\'' || s[0] == '"' {
		q, err := strconv.Unquote(s)
		if err != nil {
			errorf("invalid token: %s", err)
		}
		r := []rune(q)
		if len(r) != 1 {
			errorf("expect single character literal token, but got: %s", s)
		}
		val = int(r[0])
		if val == 0 {
			errorf("token value 0 is illegal")
		}
		terms[termN].isConst = false
	} else {
		val = extval
		extval++
		if s[0] == '$' {
			terms[termN].isConst = false
		}
	}
	terms[termN].value = val
	return termN
}

func findTerm(s string) int {
	for i, t := range terms {
		if t.name == s {
			return i
		}
	}
	return -1
}

// find term s, insert if not found
func findInsertTerm(s string) int {
	if i := findTerm(s); i > 0 {
		return i
	}
	return defineTerm(s)
}

func findNonterm(s string) int {
	for i, nt := range nonterms {
		if nt.name == s {
			return i + NTBASE
		}
	}
	return -1
}

// find nonterm s, insert if not found
func findInsertNT(s string) int {
	if i := findNonterm(s); i >= 0 {
		return i
	}
	return defineNT(s)
}

func findSym(s string) int {
	if i := findTerm(s); i > 0 {
		return i
	}
	return findNonterm(s)
}
