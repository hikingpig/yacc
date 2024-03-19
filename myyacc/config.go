package myyacc

import "flag"

const (
	SYMINC   = 50
	PRODINC  = 100    // increase for productions     prodptr
	WSETINC  = 50     // increase for working sets    wsets
	STATEINC = 200    // increase for states          statemem
	PRIVATE  = 0xE000 // value of long terminal symbol begins from unicode's private usage, except for $end
	NTBASE   = 010000 // value for non-terminal symbol starts from NTBASE
	RULEINC  = 50     // increase for max rule length prodptr[i]
	TEMPSIZE = 16000
	TOKSTART = 4 //index of first defined token
	NSTATES  = 16000
	ACTSIZE  = 240000
)

var oflag string  // go source file
var vflag string  // output file
var lflag bool    // -l disable line directives
var prefix string // name prefix for identifiers, default yy

func init() {
	flag.StringVar(&oflag, "o", "y.go", "parser output")
	flag.StringVar(&vflag, "v", "y.output", "create parsing tables")
	flag.BoolVar(&lflag, "l", false, "disable line directives")
	flag.StringVar(&prefix, "p", "yy", "name prefix to use in generated code")
}
