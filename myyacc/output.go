package myyacc

import "fmt"

func arrayOutColumns(s string, v []int, columns int, allowUnsigned bool) {
	s = prefix + s
	ftable.WriteRune('\n')
	minType := minType(v, allowUnsigned)
	fmt.Fprintf(ftable, "var %v = [...]%s{", s, minType)
	for i, val := range v {
		if i%columns == 0 {
			fmt.Fprintf(ftable, "\n\t")
		} else {
			ftable.WriteRune(' ')
		}
		fmt.Fprintf(ftable, "%d,", val)
	}
	fmt.Fprintf(ftable, "\n}\n")
}

func arout(s string, v []int, n int) {
	arrayOutColumns(s, v[:n], 10, true)
}

// output the summary on y.output
func summary() {
	if foutput != nil {
		fmt.Fprintf(foutput, "\n%v terminals, %v nonterminals\n", nterm, nnonterm+1)
		fmt.Fprintf(foutput, "%v grammar rules, %v/%v states\n", nprod, nstate, NSTATES)
		fmt.Fprintf(foutput, "%v shift/reduce, %v reduce/reduce conflicts reported\n", zzsrconf, zzrrconf)
		fmt.Fprintf(foutput, "%v working sets used\n", len(wsets))
		fmt.Fprintf(foutput, "memory: parser %v/%v\n", memp, ACTSIZE)
		fmt.Fprintf(foutput, "%v extra closures\n", zzclose-2*nstate)
		fmt.Fprintf(foutput, "%v shift entries, %v exceptions\n", zzacent, zzexcp)
		fmt.Fprintf(foutput, "%v goto entries\n", zzgoent)
		fmt.Fprintf(foutput, "%v entries saved by goto default\n", zzgobest)
	}
	if zzsrconf != 0 || zzrrconf != 0 {
		fmt.Printf("\nconflicts: ")
		if zzsrconf != 0 {
			fmt.Printf("%v shift/reduce", zzsrconf)
		}
		if zzsrconf != 0 && zzrrconf != 0 {
			fmt.Printf(", ")
		}
		if zzrrconf != 0 {
			fmt.Printf("%v reduce/reduce", zzrrconf)
		}
		fmt.Printf("\n")
	}
}

// write optimizer summary
func osummary() {
	if foutput == nil {
		return
	}
	i := 0
	for p := maxa; p >= 0; p-- {
		if amem[p] == 0 {
			i++
		}
	}

	fmt.Fprintf(foutput, "Optimizer space used: output %v/%v\n", maxa+1, ACTSIZE)
	fmt.Fprintf(foutput, "%v table entries, %v zero\n", maxa+1, i)
	fmt.Fprintf(foutput, "maximum spread: %v, maximum offset: %v\n", maxspr, maxoff)
}

func usage() {
	fmt.Fprintf(stderr, "usage: yacc [-o output] [-v parsetable] input\n")
	exit(1)
}
