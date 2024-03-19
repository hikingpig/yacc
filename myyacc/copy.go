package myyacc

import (
	"bufio"
	"bytes"
	"fmt"
	"unicode"
)

var ftable *bufio.Writer
var fcode *bytes.Buffer
var infile string
var fmtImported bool // output file has recorded an import of "fmt"

// copy the union declaration to the output, and the define file if present
func cpyunion() {

	if !lflag {
		fmt.Fprintf(ftable, "\n//line %v:%v\n", infile, lineno)
	}
	fmt.Fprintf(ftable, "type %sSymType struct", prefix)

	level := 0

out:
	for {
		c := getrune(finput)
		if c == EOF {
			errorf("EOF encountered while processing %%union")
		}
		ftable.WriteRune(c)
		switch c {
		case '\n':
			lineno++
		case '{':
			if level == 0 {
				fmt.Fprintf(ftable, "\n\tyys int")
			}
			level++
		case '}':
			level--
			if level == 0 {
				break out
			}
		}
	}
	fmt.Fprintf(ftable, "\n\n")
}

// break code into lines
func lines(code []rune) [][]rune {
	l := make([][]rune, 0, 100)
	for len(code) > 0 {
		// one line per loop
		var i int
		for i = range code {
			if code[i] == '\n' { //========== nice, i has a value
				break
			}
		}
		l = append(l, code[:i+1])
		code = code[i+1:]
	}
	return l
}

// does this line look like a package clause?  not perfect: might be confused by early comments.
func isPackageClause(line []rune) bool {
	line = skipspace(line)

	// must be big enough.
	if len(line) < len("package X\n") {
		return false
	}

	// must start with "package"
	for i, r := range []rune("package") {
		if line[i] != r {
			return false
		}
	}
	line = skipspace(line[len("package"):])

	// must have another identifier.
	if len(line) == 0 || (!unicode.IsLetter(line[0]) && line[0] != '_') {
		return false
	}
	for len(line) > 0 {
		if !unicode.IsLetter(line[0]) && !unicode.IsDigit(line[0]) && line[0] != '_' {
			break
		}
		line = line[1:]
	}
	line = skipspace(line)

	// eol, newline, or comment must follow
	if len(line) == 0 {
		return true
	}
	if line[0] == '\r' || line[0] == '\n' {
		return true
	}
	if len(line) >= 2 {
		return line[0] == '/' && (line[1] == '/' || line[1] == '*')
	}
	return false
}

// skip initial spaces
func skipspace(line []rune) []rune {
	for len(line) > 0 {
		if line[0] != ' ' && line[0] != '\t' {
			break
		}
		line = line[1:]
	}
	return line
}

// emits code saved up from between %{ and %}
// called by cpycode
// adds an import for __yyfmt__ after the package clause
func emitcode(code []rune, lineno int) {
	for i, line := range lines(code) {
		writecode(line)
		if !fmtImported && isPackageClause(line) {
			fmt.Fprintln(ftable, `import __yyfmt__ "fmt"`)
			if !lflag {
				fmt.Fprintf(ftable, "//line %v:%v\n\t\t", infile, lineno+i)
			}
			fmtImported = true
		}
	}
}

// writes code to ftable
func writecode(code []rune) {
	for _, r := range code {
		ftable.WriteRune(r)
	}
}

// saves code between %{ and %}
// adds an import for __fmt__ the first time
func cpycode() {
	lno := lineno

	c := getrune(finput)
	if c == '\n' {
		c = getrune(finput)
		lineno++
	}
	if !lflag {
		fmt.Fprintf(ftable, "\n//line %v:%v\n", infile, lineno)
	}
	// accumulate until %}
	code := make([]rune, 0, 1024)
	for c != EOF {
		if c == '%' {
			c = getrune(finput)
			if c == '}' {
				emitcode(code, lno+1)
				return
			}
			code = append(code, '%')
		}
		code = append(code, c)
		if c == '\n' {
			lineno++
		}
		c = getrune(finput)
	}
	lineno = lno
	errorf("eof before %%}")
}

// copy action to the next ; or closing }
func cpyact(curprod []int, max int) {

	if !lflag {
		fmt.Fprintf(fcode, "\n//line %v:%v", infile, lineno)
	}
	fmt.Fprint(fcode, "\n\t\t")

	lno := lineno
	brac := 0

loop:
	for {
		c := getrune(finput)

	swt:
		switch c {
		case ';':
			if brac == 0 {
				fcode.WriteRune(c)
				return
			}

		case '{':
			brac++

		case '$':
			s := 1
			tok := -1
			c = getrune(finput)

			// type description
			if c == '<' { // ======== it doesn't handle $<int>$1 properly, neither $<int>$$
				ungetrune(finput, c)
				if gettok() != TYPENAME {
					errorf("bad syntax on $<ident> clause")
				}
				tok = numbval
				c = getrune(finput)
			}
			if c == '$' { // ========= $$ ->yyVAL, $<int>$ -> yyVAL.int. the logic seems flaw.
				fmt.Fprintf(fcode, "%sVAL", prefix)

				// put out the proper tag...
				if ntypes != 0 {
					if tok < 0 {
						tok = symType(curprod[0])
					}
					fmt.Fprintf(fcode, ".%v", typeset[tok])
				}
				continue loop
			}
			if c == '-' {
				s = -s
				c = getrune(finput)
			}
			j := 0
			if isdigit(c) {
				for isdigit(c) {
					j = j*10 + int(c-'0')
					c = getrune(finput)
				}
				ungetrune(finput, c)
				j = j * s
				if j >= max {
					errorf("Illegal use of $%v", j)
				}
			} else if isword(c) || c == '.' {
				// look for $name
				ungetrune(finput, c)
				if gettok() != IDENTIFIER {
					errorf("$ must be followed by an identifier")
				}
				tokn := chfind(2, tokname)
				fnd := -1
				c = getrune(finput)
				if c != '@' {
					ungetrune(finput, c)
				} else if gettok() != NUMBER {
					errorf("@ must be followed by number")
				} else {
					fnd = numbval
				}
				for j = 1; j < max; j++ {
					if tokn == curprod[j] {
						fnd--
						if fnd <= 0 {
							break
						}
					}
				}
				if j >= max {
					errorf("$name or $name@number not found")
				}
			} else {
				fcode.WriteRune('$')
				if s < 0 {
					fcode.WriteRune('-')
				}
				ungetrune(finput, c)
				continue loop
			}
			fmt.Fprintf(fcode, "%sDollar[%v]", prefix, j)

			// put out the proper tag
			if ntypes != 0 {
				if j <= 0 && tok < 0 { // this can occur normally. we better not supporting negative number ?
					errorf("must specify type of $%v", j)
				}
				if tok < 0 {
					tok = symType(curprod[j])
				}
				fmt.Fprintf(fcode, ".%v", typeset[tok])
			}
			continue loop

		case '}':
			brac--
			if brac != 0 { // ============= why break?
				break
			}
			fcode.WriteRune(c)
			return

		case '/': // ========== how it handle comment?
			nc := getrune(finput)
			if nc != '/' && nc != '*' {
				ungetrune(finput, nc)
				break
			}
			// a comment
			fcode.WriteRune(c)
			fcode.WriteRune(nc)
			c = getrune(finput)
			for c != EOF {
				switch {
				case c == '\n':
					lineno++
					if nc == '/' { // end of // comment
						break swt // ====== break swt, go to end of loop, write rune
					}
				case c == '*' && nc == '*': // end of /* comment?
					nnc := getrune(finput)
					if nnc == '/' {
						fcode.WriteRune('*')
						fcode.WriteRune('/')
						continue loop // ======== start at the beginning, not write rune anymore!
					}
					ungetrune(finput, nnc)
				}
				fcode.WriteRune(c)
				c = getrune(finput)
			}
			errorf("EOF inside comment")

		case '\'', '"':
			// character string or constant
			match := c
			fcode.WriteRune(c)
			c = getrune(finput)
			for c != EOF {
				if c == '\\' {
					fcode.WriteRune(c)
					c = getrune(finput)
					if c == '\n' {
						lineno++
					}
				} else if c == match {
					break swt // ========= break and write the rune at the end
				}
				if c == '\n' {
					errorf("newline in string or char const")
				}
				fcode.WriteRune(c)
				c = getrune(finput)
			}
			errorf("EOF in string or character constant")

		case EOF:
			lineno = lno
			errorf("action does not terminate")

		case '\n':
			fmt.Fprint(fcode, "\n\t")
			lineno++
			continue loop
		}

		fcode.WriteRune(c)
	}
}
