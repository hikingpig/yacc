package myyacc

import (
	"fmt"
	"strings"
)

var nolook = 0 // flag to turn off lookahead computations

/*
Lkset implements Lookahead set for LR(1) parsing.
Each integer in the set represents 32 bits.
In the set, the order of integers are from left to right
but in each integer, the order of bits are from right to left.
For example, 1st bit is in the 1st integer, but it is the last bit of that integer!
*/
type Lkset []int

/*
set sets a bit at position `bit` to 1
bit>>5 retrieves the integer for `bit`
bit&31 takes the last 5 bits of `bit`, essentially the same as `bit` if it is less than 5 bits
a |= (1<<b) set the bit bth from the END of a
*/
func (s Lkset) set(bit int) {
	s[bit>>5] |= (1 << uint(bit&31))
}

/*
isSet checks if bit `bit` is ON in the set
*/
func (s Lkset) check(bit int) bool {
	return s[bit>>5]&(1<<uint(bit&31)) == 0
}

/*
union does the union of set s with r.
the size of union is size of s
if set s changed, return true
*/
func (s Lkset) union(r Lkset) bool {
	changed := false
	for i := 0; i < min(len(s), len(r)); i++ {
		if s[i] != r[i] {
			changed = true
			s[i] |= r[i]
		}
	}
	return changed
}

/*
String prints the list of tokens in Lkset
*/
func (s Lkset) String() string {
	if s == nil {
		return fmt.Sprint("\tNULL")

	}
	buf := strings.Builder{}
	buf.WriteString(" { ")
	for i := 0; i <= nterm; i++ {
		if s.check(i) {
			buf.WriteString(terms[i].name)
		}
	}
	buf.WriteString("}")
	return buf.String()
}

func newLkset() Lkset {
	return make([]int, tbitset)
}
