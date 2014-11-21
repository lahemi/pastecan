package pbnf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"unicode"
)

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
	fgDir = os.Getenv("HOME") + "/.local/share/pastecan/fgs/"

	basesyntax = genSyntax("base")
	isSpace    = isGen(basesyntax["spaces"])
	isPunct    = isGen(basesyntax["puncts"])
	htmlescs   = map[string]string{
		">": "&gt;", "<": "&lt;", "&": "&amp;",
	}
)

func stderr(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a)
}
func dierr(emsg ...interface{}) {
	stderr(emsg)
	os.Exit(1)
}

func escape(s string, escapes map[string]string) string {
	for k, v := range escapes {
		if s == k {
			return v
		}
	}
	return s
}

// Need a way to do repeats in `grammar|syntax` files.
func genSyntax(lang string) map[string][]string {
	cnt, err := ioutil.ReadFile(fgDir + lang + ".fg")
	if err != nil {
		dierr(err)
	}
	// Crude, could and should be dealt with more elegantly.
	conv := func(str string) string {
		escs := map[string]string{
			`\`: " ", `\n`: "\n", `\r`: "\r",
			`\t`: "\t", `\;`: ";", `\\`: `\`,
		}
		return escape(str, escs)
	}

	const (
		RD = iota
		IDENT
		COMP
		CMT
	)
	// These are runes for simplicity.
	const (
		cmtMark   = '#'
		identMark = ':'
		endMark   = ';'
	)
	var (
		retMap = make(map[string][]string)
		buf    string
		ident  string
		bufs   []string
		text   = []rune(string(cnt))
		state  = RD
	)

	for i := 0; i < len(text); i++ {
		c := text[i]
		switch state {
		case CMT:
			if c == '\n' {
				state = RD
			}

		case RD:
			switch {
			case c == cmtMark:
				state = CMT
			case c == identMark:
				state = IDENT
			}

		case IDENT:
			switch {
			case unicode.IsSpace(c) && buf == "":

			case unicode.IsSpace(c) && buf != "":
				ident = buf
				buf = ""
				state = COMP

			default:
				buf += string(c)
			}

		case COMP:
			switch {
			case unicode.IsSpace(c) && buf == "":

			case unicode.IsSpace(c) && buf != "":
				bufs = append(bufs, conv(buf))
				buf = ""

			case c == endMark && buf == "":
				retMap[ident] = bufs

				bufs = []string{}
				ident = ""
				state = RD

			default:
				buf += string(c)
			}
		}
	}

	return retMap
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

func inferSyntax(lang string) (map[string][]string, error) {
	switch lang {
	case "go":
		return genSyntax("go"), nil
	case "lua":
		return genSyntax("lua"), nil
	default:
		return nil, errors.New("no syntax file found")
	}
}

// This is by far the most awful part of this code.
// It is also somewhat incomplete and potentially buggy.
func Colourify(lang, body string) string {
	syntax, err := inferSyntax(lang)
	if err != nil {
		dierr(err)
	}
	const (
		RD = iota
		SGLCMT
		MULCMT
		STR
	)
	var (
		output               []string
		strcmp, word, outstr string

		isStrChar = isGen(syntax["strchars"])
		isKeyword = isGen(syntax["keywords"])
		isBuiltin = isGen(syntax["builtins"])

		text  = []rune(body)
		state = RD
	)

	out := func(sp ...string) {
		output = append(output, sp...)
	}
	pf := func(ps string) {
		p := escape(ps, htmlescs)
		if isSpace(p) || isPunct(p) {
			switch {
			case state == RD && isKeyword(word):
				out(keyspan, word, es, p)
			case state == RD && isBuiltin(word):
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

	for i := 0; i < len(text); i++ {
		c := string(text[i])
		switch state {
		case RD:
			switch {
			case isStrChar(c):
				out(word, strspan)
				state = STR
				strcmp = c
				word = ""
			case i < len(text)-1 && syntest(c+string(text[i+1]), "mulcmts"):
				out(cmtspan)
				state = MULCMT
			case i < len(text)-1 && syntest(c+string(text[i+1]), "comment"):
				out(cmtspan)
				state = SGLCMT
			}

		case MULCMT:
			if syntest(string(text[i-1])+c, "mulcmte") {
				out(word, c, es)
				state = RD
				c, word = "", ""
			}

		case STR:
			if isStrChar(c) && strcmp == c {
				out(word, c, es)
				state = RD
				c, word, strcmp = "", "", ""
			}

		case SGLCMT:
			if c == "\n" {
				out(es)
				state = RD
				c, word = "", ""
			}
		}
		pf(c)
	}

	for _, v := range output {
		outstr += v
	}
	return outstr
}
