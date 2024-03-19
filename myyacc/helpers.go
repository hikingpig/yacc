package myyacc

func extend[T Symbol | int](s *[]T, INC int) {
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
