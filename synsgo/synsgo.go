/*
A very bad Golang parser, mashing span tags at appropriate
points. You shouldn't probably use this. Not yet at least.
TODO: Make keywords and string chars and such externally
configurable, so as to not require recompilation everytime.
AGPL, 2014, Lauri Peltomäki
*/
package synsgo

import (
	"strings"
)

var keywords = []string{
	"break", "default", "func", "interface", "select",
	"case", "defer", "go", "map", "struct",
	"chan", "else", "goto", "package", "switch",
	"const", "fallthrough", "if", "range", "type",
	"continue", "for", "import", "return", "var",
}

var cmtspan = `<span class="comment">`
var strspan = `<span class="string">`
var keyspan = `<span class="keyword">`
var es = `</span>`

var incmt = false
var inmul = false
var instr = false

func isStrChar(s string) bool {
	if s == `"` || s == `'` || s == "`" {
		return true
	}
	return false
}
func isSpace(s string) bool {
	if s == " " || s == "\t" || s == "\n" {
		return true
	}
	return false
}
func isKeyword(s string) bool {
	for _, v := range keywords {
		if s == v {
			return true
		}
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

func Colourify(body string) string {
	s := strings.Split(body, "")
	var strtmp, word string
	var output []string
	for i := 0; i < len(s); i++ {
		switch {
		case !instr && !incmt && isStrChar(s[i]):
			output = append(output, word, strspan, s[i])
			instr = true
			strtmp = s[i]
			word = ""
		case instr && isStrChar(s[i]):
			if strtmp == s[i] {
				output = append(output, word, s[i], es)
				instr = false
				word, strtmp = "", ""
			}
		case !incmt && !instr && s[i] == "/":
			if s[i+1] == "/" || s[i+1] == "*" {
				output = append(output, cmtspan)
				incmt = true
				if s[i+1] == "*" {
					inmul = true
				}
			}
			output = append(output, s[i])
		case inmul && s[i] == "/":
			if s[i-1] == "*" {
				inmul, incmt = false, false
				output = append(output, word, s[i], es)
				word = ""
			} else {
				output = append(output, s[i])
			}
		case incmt && !inmul && s[i] == "\n":
			output = append(output, word, es)
			word = ""
			incmt = false
		default:
			p := htmlesc(s[i])
			if isSpace(p) {
				if isKeyword(word) && !inmul && !incmt {
					output = append(output, keyspan, word, es, p)
				} else {
					output = append(output, word, p)
				}
				word = ""
			} else {
				word += p
			}
		}
	}
	// Seems a bit nasty to me... Also bad should the file be large.
	var out string
	for _, v := range output {
		out += v
	}
	return out
}
