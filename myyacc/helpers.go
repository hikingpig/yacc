package myyacc

import "math"

func extend[T any](s *[]T, INC int) {
	new := make([]T, len(*s)+INC)
	copy(new, *s)
	*s = new
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func assert(cond bool, msgFormat string, v ...interface{}) {
	if !cond {

	}
}

func TYPE(i int) int       { return (i >> 10) & 077 }
func SETTYPE(i, j int) int { return i | (j << 10) }

// macros for getting associativity and precedence levels
func ASSOC(i int) int { return i & 3 }

// macros for setting associativity and precedence levels
func SETASC(i, j int) int  { return i | j }
func SETPLEV(i, j int) int { return i | (j << 4) }
func PLEVEL(i int) int     { return (i >> 4) & 077 }

func isdigit(c rune) bool { return c >= '0' && c <= '9' }

func isword(c rune) bool {
	return c >= 0xa0 || c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// set elements 0 through n-1 to c
func aryfil(v []int, n, c int) {
	for i := 0; i < n; i++ {
		v[i] = c
	}
}

// return 1 if 2 arrays are equal
// return 0 if not equal
func aryeq(a []int, b []int) int {
	n := len(a)
	if len(b) != n {
		return 0
	}
	for ll := 0; ll < n; ll++ {
		if a[ll] != b[ll] {
			return 0
		}
	}
	return 1
}

// copies and protects "'s in q
func chcopy(q string) string {
	s := ""
	i := 0
	j := 0
	for i = 0; i < len(q); i++ {
		if q[i] == '"' {
			s += q[j:i] + "\\"
			j = i
		}
	}
	return s + q[j:i]
}

func minMax(v []int) (min, max int) {
	if len(v) == 0 {
		return
	}
	min = v[0]
	max = v[0]
	for _, i := range v {
		if i < min {
			min = i
		}
		if i > max {
			max = i
		}
	}
	return
}

// return the smaller integral base type to store the values in v
func minType(v []int, allowUnsigned bool) (typ string) {
	typ = "int"
	typeLen := 8
	min, max := minMax(v)
	checkType := func(name string, size, minType, maxType int) {
		if min >= minType && max <= maxType && typeLen > size {
			typ = name
			typeLen = size
		}
	}
	checkType("int32", 4, math.MinInt32, math.MaxInt32)
	checkType("int16", 2, math.MinInt16, math.MaxInt16)
	checkType("int8", 1, math.MinInt8, math.MaxInt8)
	if allowUnsigned {
		// Do not check for uint32, not worth and won't compile on 32 bit systems
		checkType("uint16", 2, 0, math.MaxUint16)
		checkType("uint8", 1, 0, math.MaxUint8)
	}
	return
}
