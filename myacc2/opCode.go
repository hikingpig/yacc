package myacc2

// i compacts info of associativity, precedence, type of a production or terminal symbol
func asc(i int) int { return i & 3 } // last 2 bits

func prec(i int) int { return (i >> 4) & 077 } // skip 4 bits, use 6 bits

func typ(i int) int { return (i >> 10) & 077 } // skip 10 bits, use 6 bits

func addAsc(i, j int) int { return i | j } // j <=3

func addPrec(i, j int) int { return i | (j << 4) } // j<= 077

func addTyp(i, j int) int { return i | (j << 10) } // j <= 077
