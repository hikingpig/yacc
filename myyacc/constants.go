package myyacc

// flags for a rule having an action, and being reduced
const (
	ACTFLAG = 1 << (iota + 2)
	REDFLAG
)
const (
	NTBASE     = 010000
	TOKSTART   = 4 //index of first defined token
	ERRCODE    = 8190
	ACCEPTCODE = 8191
	YYLEXUNK   = 3
)
