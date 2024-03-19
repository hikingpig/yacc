package myyacc

// flags for a rule having an action, and being reduced
const (
	ACTFLAG = 1 << (iota + 2)
	REDFLAG
)
