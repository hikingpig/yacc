package myacc2

import (
	"bufio"
	"flag"
	"os"
)

var oflag string  // -o [y.go] -- file to store ftable
var vflag string  // -v [y.output]	-- file to store foutput
var prefix string // name prefix for identifiers, default yy

var infile string         // name of input file
var finput *bufio.Reader  // input file
var ftable *bufio.Writer  // the generated go source file
var foutput *bufio.Writer // file to write state output

func init() {
	flag.StringVar(&oflag, "o", "y.go", "parser output")
	flag.StringVar(&prefix, "p", "yy", "name prefix to use in generated code")
	flag.StringVar(&vflag, "v", "y.output", "create parsing tables")
}

// open input file to read and output files to write to
func openup() {
	infile = flag.Arg(0)
	finput = open(infile)
	if finput == nil {
		errorf("cannot open %v", infile)
	}

	foutput = nil
	foutput = create(vflag)
	if foutput == nil {
		errorf("can't create file %v", vflag)
	}

	ftable = nil
	ftable = create(oflag)
	if ftable == nil {
		errorf("can't create file %v", oflag)
	}

}

// open input file to read
func open(s string) *bufio.Reader {
	fi, err := os.Open(s)
	if err != nil {
		errorf("error opening %v: %v", s, err)
	}
	//fmt.Printf("open %v\n", s);
	return bufio.NewReader(fi)
}

// create file to write (foutput, ftable)
func create(s string) *bufio.Writer {
	fo, err := os.Create(s)
	if err != nil {
		errorf("error creating %v: %v", s, err)
	}
	//fmt.Printf("create %v mode %v\n", s);
	return bufio.NewWriter(fo)
}
