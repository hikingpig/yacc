package myyacc

var yaccpar string // will be processed version of yaccpartext: s/$$/prefix/g
var template = `
/*	parser for yacc output	*/

var (
	$$Debug        = 0
	$$ErrorVerbose = false
)

type $$Lexer interface {
	Lex(lval *$$SymType) int
	Error(s string)
}

type $$Parser interface {
	Parse($$Lexer) int
	Lookahead() int
}

type $$ParserImpl struct {
	lval  $$SymType
	stack [$$InitialStackSize]$$SymType
	char  int
}

func (p *$$ParserImpl) Lookahead() int {
	return p.char
}

func $$NewParser() $$Parser {
	return &$$ParserImpl{}
}

const $$Flag = -1000

func $$Tokname(c int) string {
	if c >= 1 && c-1 < len($$Toknames) {
		if $$Toknames[c-1] != "" {
			return $$Toknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func $$Statname(s int) string {
	if s >= 0 && s < len($$Statenames) {
		if $$Statenames[s] != "" {
			return $$Statenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func $$ErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !$$ErrorVerbose {
		return "syntax error"
	}

	for _, e := range $$ErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + $$Tokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int($$Pact[state])
	for tok := TOKSTART; tok-1 < len($$Toknames); tok++ {
		if n := base + tok; n >= 0 && n < $$Last && int($$Chk[int($$Act[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if $$Def[state] == -2 {
		i := 0
		for $$Exca[i] != -1 || int($$Exca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; $$Exca[i] >= 0; i += 2 {
			tok := int($$Exca[i])
			if tok < TOKSTART || $$Exca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if $$Exca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += $$Tokname(tok)
	}
	return res
}

func $$lex1(lex $$Lexer, lval *$$SymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int($$Tok1[0])
		goto out
	}
	if char < len($$Tok1) {
		token = int($$Tok1[char])
		goto out
	}
	if char >= $$Private {
		if char < $$Private+len($$Tok2) {
			token = int($$Tok2[char-$$Private])
			goto out
		}
	}
	for i := 0; i < len($$Tok3); i += 2 {
		token = int($$Tok3[i+0])
		if token == char {
			token = int($$Tok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int($$Tok2[1]) /* unknown char */
	}
	if $$Debug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", $$Tokname(token), uint(char))
	}
	return char, token
}

func $$Parse($$lex $$Lexer) int {
	return $$NewParser().Parse($$lex)
}

func ($$rcvr *$$ParserImpl) Parse($$lex $$Lexer) int {
	var $$n int
	var $$VAL $$SymType
	var $$Dollar []$$SymType
	_ = $$Dollar // silence set and not used
	$$S := $$rcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	$$state := 0
	$$rcvr.char = -1
	$$token := -1 // $$rcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		$$state = -1
		$$rcvr.char = -1
		$$token = -1
	}()
	$$p := -1
	goto $$stack

ret0:
	return 0

ret1:
	return 1

$$stack:
	/* put a state and value onto the stack */
	if $$Debug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", $$Tokname($$token), $$Statname($$state))
	}

	$$p++
	if $$p >= len($$S) {
		nyys := make([]$$SymType, len($$S)*2)
		copy(nyys, $$S)
		$$S = nyys
	}
	$$S[$$p] = $$VAL
	$$S[$$p].yys = $$state

$$newstate:
	$$n = int($$Pact[$$state])
	if $$n <= $$Flag {
		goto $$default /* simple state */
	}
	if $$rcvr.char < 0 {
		$$rcvr.char, $$token = $$lex1($$lex, &$$rcvr.lval)
	}
	$$n += $$token
	if $$n < 0 || $$n >= $$Last {
		goto $$default
	}
	$$n = int($$Act[$$n])
	if int($$Chk[$$n]) == $$token { /* valid shift */
		$$rcvr.char = -1
		$$token = -1
		$$VAL = $$rcvr.lval
		$$state = $$n
		if Errflag > 0 {
			Errflag--
		}
		goto $$stack
	}

$$default:
	/* default state action */
	$$n = int($$Def[$$state])
	if $$n == -2 {
		if $$rcvr.char < 0 {
			$$rcvr.char, $$token = $$lex1($$lex, &$$rcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if $$Exca[xi+0] == -1 && int($$Exca[xi+1]) == $$state {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			$$n = int($$Exca[xi+0])
			if $$n < 0 || $$n == $$token {
				break
			}
		}
		$$n = int($$Exca[xi+1])
		if $$n < 0 {
			goto ret0
		}
	}
	if $$n == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			$$lex.Error($$ErrorMessage($$state, $$token))
			Nerrs++
			if $$Debug >= 1 {
				__yyfmt__.Printf("%s", $$Statname($$state))
				__yyfmt__.Printf(" saw %s\n", $$Tokname($$token))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for $$p >= 0 {
				$$n = int($$Pact[$$S[$$p].yys]) + $$ErrCode
				if $$n >= 0 && $$n < $$Last {
					$$state = int($$Act[$$n]) /* simulate a shift of "error" */
					if int($$Chk[$$state]) == $$ErrCode {
						goto $$stack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if $$Debug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", $$S[$$p].yys)
				}
				$$p--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if $$Debug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", $$Tokname($$token))
			}
			if $$token == $$EofCode {
				goto ret1
			}
			$$rcvr.char = -1
			$$token = -1
			goto $$newstate /* try again in the same state */
		}
	}

	/* reduction by production $$n */
	if $$Debug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", $$n, $$Statname($$state))
	}

	$$nt := $$n
	$$pt := $$p
	_ = $$pt // guard against "declared and not used"

	$$p -= int($$R2[$$n])
	// $$p is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if $$p+1 >= len($$S) {
		nyys := make([]$$SymType, len($$S)*2)
		copy(nyys, $$S)
		$$S = nyys
	}
	$$VAL = $$S[$$p+1]

	/* consult goto table to find next state */
	$$n = int($$R1[$$n])
	$$g := int($$Pgo[$$n])
	$$j := $$g + $$S[$$p].yys + 1

	if $$j >= $$Last {
		$$state = int($$Act[$$g])
	} else {
		$$state = int($$Act[$$j])
		if int($$Chk[$$state]) != -$$n {
			$$state = int($$Act[$$g])
		}
	}
	// dummy call; replaced with literal code
	$$run()
	goto $$stack /* stack new state and value */
}
`
