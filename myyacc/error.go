package myyacc

import "fmt"

var nerrors = 0 // number of errors
var fatfl = 1   // if on, error is fatal

func errorf(s string, v ...interface{}) {} // use fmt.Errorf?
// write out error comment
func lerrorf(lineno int, s string, v ...interface{}) {
	nerrors++
	fmt.Fprintf(stderr, s, v...)
	fmt.Fprintf(stderr, ": %v:%v\n", infile, lineno)
	if fatfl != 0 {
		summary()
		exit(1)
	}
}

type Error struct {
	lineno int
	tokens []string
	msg    string
}

var errors []Error
