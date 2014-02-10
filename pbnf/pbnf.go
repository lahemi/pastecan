package pbnf

import (
	"io/ioutil"
	"os"
	"strings"
)

type Slm struct {
	Name  string
	Elems []string
}

/*
<rule>      ::= <text> <opt-white> ":=" <opt-white> <expr> <EOL>
<opt-white> ::= " " <opt-white> | ""
<expr>      ::= <element> | <element> "|" <expr>
<element>   ::= '"' <text> '"'

! need a way to denote 0 or more repeats
*/
func genSyntax(lang string) map[string][]string {
	s, e := ioutil.ReadFile(os.Getenv("HOME") + "/.local/share/pastecan/fgs/" + lang + ".fg")
	if e != nil {
		panic(e)
	}
	// Crude, could and should be dealt with more elegantly.
	conv := func(str string) string {
		switch {
		case str == `\n`:
			str = "\n"
		case str == `\r`:
			str = "\r"
		case str == `\t`:
			str = "\t"
		}
		return str
	}

	var (
		sls   = make(map[string][]string)
		slm   Slm
		t, ts string

		inSym, inExpr = true, false
	)

	for i := 0; i < len(s); i++ {
		switch {
		case inSym:
			if s[i] == ' ' {
				continue
			}
			if s[i] == ':' && s[i+1] == '=' {
				slm.Name = ts
				ts = ""
				inSym = false
			} else {
				ts += string(s[i])
			}
		case s[i] == '|' && !inExpr:
			t = conv(t)
			slm.Elems = append(slm.Elems, t)
			t = ""
		case s[i] == '\n' && !inExpr:
			if s[i-1] != '|' && s[i-1] != ',' {
				inSym = true
			}
		case (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') && !inExpr:
		case !inExpr && s[i] == '"':
			inExpr = true
		case inExpr:
			if s[i] == '"' {
				inExpr = false
				if s[i+1] == '\n' {
					t = conv(t)
					slm.Elems = append(slm.Elems, t)
					sls[slm.Name] = slm.Elems
					t = ""
					slm.Elems = []string{}
				}
				if s[i+1] == '"' {
					inExpr = true
				} else {
					continue
				}
			}
			t += string(s[i])
		}
	}

	return sls
}

func isGen(l []string) func(string) bool {
	return func(s string) bool {
		for _, v := range l {
			if s == v {
				return true
			}
		}
		return false
	}
}

const (
	cmtspan = `<span class="comment">`
	strspan = `<span class="string">`
	keyspan = `<span class="keyword">`
	bltspan = `<span class="builtin">`
	es      = `</span>`
)

// In string, in multiline comment, in normal comment.
// Like the octal value in Unix permissions.
var (
	stateval   = 0
	basesyntax = genSyntax("base")
	isSpace    = isGen(basesyntax["spaces"])
	isPunct    = isGen(basesyntax["puncts"])
)

func state(st int) bool {
	if stateval == st {
		return true
	}
	return false
}

var escapes = map[string]string{
	">": "&gt;", "<": "&lt;", "&": "&amp;",
}

func htmlesc(s string) string {
	for k, v := range escapes {
		if s == k {
			return v
		}
	}
	return s
}

func Colourify(lang, body string) string {
	var syntax map[string][]string
	switch lang {
	case "go":
		syntax = genSyntax("go")
	case "lua":
		syntax = genSyntax("lua")
	default:
		return ""
	}
	var (
		output               []string
		strtmp, word, outstr string

		isStrChar = isGen(syntax["strchars"])
		isKeyword = isGen(syntax["keywords"])
		isBuiltin = isGen(syntax["builtins"])
	)

	out := func(sp ...string) {
		output = append(output, sp...)
	}
	pf := func(ps string) {
		p := htmlesc(ps)
		if isSpace(p) || isPunct(p) {
			switch {
			case state(0) && isKeyword(word):
				out(keyspan, word, es, p)
			case state(0) && isBuiltin(word):
				out(bltspan, word, es, p)
			default:
				out(word, p)
			}
			word = ""
		} else {
			word += p
		}
	}
	// func(s[i]+s[i+1], key)
	syntest := func(str, key string) bool {
		if str == syntax[key][0] {
			return true
		}
		return false
	}

	s := strings.Split(body, "") // So that we'll have UTF-8 chars.
	for i := 0; i < len(s); i++ {
		switch {
		case state(0):
			switch {
			case isStrChar(s[i]):
				out(word, strspan, s[i])
				stateval += 4
				strtmp = s[i]
				word = ""
			case i < len(s)-1 && syntest(s[i]+s[i+1], "mulcmts"):
				out(cmtspan, s[i])
				stateval += 2
			case i < len(s)-1 && syntest(s[i]+s[i+1], "comment"):
				out(cmtspan, s[i])
				stateval += 1
			default:
				pf(s[i])
			}
		case state(4) && isStrChar(s[i]) && strtmp == s[i]:
			out(word, s[i], es)
			stateval -= 4
			word, strtmp = "", ""
		case state(2) && syntest(s[i-1]+s[i], "mulcmte"):
			out(word, s[i], es)
			stateval -= 2
			word = ""
		case state(1) && s[i] == "\n":
			out(word, es)
			stateval -= 1
			word = ""
		default:
			pf(s[i])
		}
	}

	for _, v := range output {
		outstr += v
	}
	return outstr
}
