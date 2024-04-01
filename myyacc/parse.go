package myyacc

import (
	"bufio"
	"fmt"
)

const (
	IDENTIFIER = PRIVATE + iota
	MARK
	TERM
	LEFT
	RIGHT
	BINARY
	PREC
	LCURLY
	IDENTCOLON
	NUMBER
	START
	TYPEDEF
	TYPENAME
	UNION
	ERROR
)

const ENDFILE = 0
const EMPTY = 1
const WHOKNOWS = 0
const OK = 1
const NOMORE = -1000
const EOF = -1

var tokname string
var lineno int

var finput *bufio.Reader // input file
var peekline = 0
var peekrune rune
var tokflag = false                // for debug message
var ntypes int                     // number of types defined
var typeset = make(map[int]string) // pointers to type tags
var numbval int                    // value of an input number

type Resrv struct {
	name  string
	value int
}

var resrv = []Resrv{
	{"binary", BINARY},
	{"left", LEFT},
	{"nonassoc", BINARY},
	{"prec", PREC},
	{"right", RIGHT},
	{"start", START},
	{"term", TERM},
	{"token", TERM},
	{"type", TYPEDEF},
	{"union", UNION},
	{"struct", UNION},
	{"error", ERROR},
}

func getrune(f *bufio.Reader) rune {
	var r rune

	if peekrune != 0 {
		if peekrune == EOF {
			return EOF
		}
		r = peekrune
		peekrune = 0
		return r
	}

	c, n, err := f.ReadRune()
	if n == 0 {
		return EOF
	}
	if err != nil {
		errorf("read error: %v", err)
	}
	//fmt.Printf("rune = %v n=%v\n", string(c), n);
	return c
}

func ungetrune(f *bufio.Reader, c rune) {
	if f != finput { // seems unnecessary?
		panic("ungetc - not finput")
	}
	if peekrune != 0 {
		panic("ungetc - 2nd unget")
	}
	peekrune = c
}

// skip over comments
// skipcom is called after reading a '/'
func skipcom() int {
	c := getrune(finput)
	if c == '/' {
		for c != EOF {
			if c == '\n' {
				return 1
			}
			c = getrune(finput)
		}
		errorf("EOF inside comment")
		return 0
	}
	if c != '*' {
		errorf("illegal comment")
	}

	nl := 0 // lines skipped
	c = getrune(finput)

l1:
	switch c {
	case '*':
		c = getrune(finput)
		if c == '/' {
			break
		}
		goto l1

	case '\n':
		nl++
		fallthrough

	default:
		c = getrune(finput)
		goto l1
	}
	return nl
}

func getword(c rune) {
	tokname = ""
	for isword(c) || isdigit(c) || c == '.' || c == '$' {
		tokname += string(c)
		c = getrune(finput)
	}
	ungetrune(finput, c)
}

func gettok() int {
	var i int
	var match, c rune

	tokname = ""
	for {
		lineno += peekline
		peekline = 0
		c = getrune(finput)
		// skip all white spaces
		for c == ' ' || c == '\n' || c == '\t' || c == '\v' || c == '\r' {
			if c == '\n' {
				lineno++
			}
			c = getrune(finput)
		}

		if c != '/' {
			break
		}
		lineno += skipcom()
	}

	switch c {
	case EOF:
		if tokflag {
			fmt.Printf(">>> ENDFILE %v\n", lineno)
		}
		return ENDFILE

	case '{':
		ungetrune(finput, c)
		if tokflag {
			fmt.Printf(">>> ={ %v\n", lineno)
		}
		return '='

	case '<':
		// get, and look up, a type name (union member name)
		c = getrune(finput)
		for c != '>' && c != EOF && c != '\n' {
			tokname += string(c)
			c = getrune(finput)
		}

		if c != '>' {
			errorf("unterminated < ... > clause")
		}

		for i = 1; i <= ntypes; i++ {
			if typeset[i] == tokname {
				numbval = i
				if tokflag {
					fmt.Printf(">>> TYPENAME old <%v> %v\n", tokname, lineno)
				}
				return TYPENAME
			}
		}
		ntypes++
		numbval = ntypes
		typeset[numbval] = tokname
		if tokflag {
			fmt.Printf(">>> TYPENAME new <%v> %v\n", tokname, lineno)
		}
		return TYPENAME

	case '"', '\'':
		match = c
		tokname = string(c)
		for {
			c = getrune(finput)
			if c == '\n' || c == EOF {
				errorf("illegal or missing ' or \"")
			}
			if c == '\\' {
				tokname += string('\\')
				c = getrune(finput)
			} else if c == match {
				if tokflag {
					fmt.Printf(">>> IDENTIFIER \"%v\" %v\n", tokname, lineno)
				}
				tokname += string(c)
				return IDENTIFIER
			}
			tokname += string(c)
		}

	case '%':
		c = getrune(finput)
		switch c {
		case '%': // MARK token, %%
			if tokflag {
				fmt.Printf(">>> MARK %%%% %v\n", lineno)
			}
			return MARK
		case '=': // PREC %=
			if tokflag {
				fmt.Printf(">>> PREC %%= %v\n", lineno)
			}
			return PREC
		case '{': // block of go code %{, don't use
			if tokflag {
				fmt.Printf(">>> LCURLY %%{ %v\n", lineno)
			}
			return LCURLY
		}

		getword(c) // reserved word
		// find a reserved word
		for i := range resrv {
			if tokname == resrv[i].name {
				if tokflag {
					fmt.Printf(">>> %%%v %v %v\n", tokname,
						resrv[i].value-PRIVATE, lineno)
				}
				return resrv[i].value // ====== if the returned value of reserved word is the same for different reserved word, it seems not correct!
			}
		}
		errorf("invalid escape, or illegal reserved word: %v", tokname)

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		numbval = int(c - '0')
		for {
			c = getrune(finput)
			if !isdigit(c) {
				break
			}
			numbval = numbval*10 + int(c-'0')
		}
		ungetrune(finput, c) // ======= we can use peek? so decide to use or not using it
		if tokflag {
			fmt.Printf(">>> NUMBER %v %v\n", numbval, lineno)
		}
		return NUMBER

	default:
		if isword(c) || c == '.' || c == '$' {
			getword(c)
			break
		}
		if tokflag {
			fmt.Printf(">>> OPERATOR %v %v\n", string(c), lineno)
		}
		return int(c)
	}

	// look ahead to distinguish IDENTIFIER from IDENTCOLON
	c = getrune(finput)
	for c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\r' || c == '/' {
		if c == '\n' {
			peekline++
		}
		// look for comments
		if c == '/' {
			peekline += skipcom()
		}
		c = getrune(finput)
	}
	if c == ':' {
		if tokflag {
			fmt.Printf(">>> IDENTCOLON %v: %v\n", tokname, lineno)
		}
		return IDENTCOLON
	}

	ungetrune(finput, c)
	if tokflag {
		fmt.Printf(">>> IDENTIFIER %v %v\n", tokname, lineno)
	}
	return IDENTIFIER
}
