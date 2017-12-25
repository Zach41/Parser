package parser

import (
	"errors"
	"unicode"
	"unicode/utf8"
)

const (
	emptyTok = iota // not used
	nonterm
	term
	literal
	newline
	begindef
	enddef
	alternate
	code
	hfield
	separate
	other
)

type TokType int

type Scanner struct {
	content []byte
	index   int
}

type WordTok struct {
	tokType TokType
	text    string
}

func (self *Scanner) Reminder() []byte {
	return self.content[self.index:]
}

func (self *Scanner) NextWord() (err error, word WordTok) {
	if self.index >= len(self.content) {
		err = errors.New("End of File")
		return
	}

Omitspace: // omit spaces
	for {
		r, l := utf8.DecodeRune(self.content[self.index:])
		if r == utf8.RuneError {
			err = errors.New("Invalid utf8 encoding")
			return
		}
		if !unicode.IsSpace(r) || r == '\n' {
			break
		}
		self.index += l
	}

	r, l := utf8.DecodeRune(self.content[self.index:])

	// omit comments
	if r == '#' {
		// comments
		for {
			self.index += l
			r, l = utf8.DecodeRune(self.content[self.index:])
			if r == utf8.RuneError {
				err = errors.New("Invalid utf8 encoding")
				return
			}
			if r == '\n' {
				goto Omitspace
			}
		}
	}

	start, inchar, incode, tokType := self.index, false, 0, other

Loop:
	for {
		if self.index >= len(self.content) {
			break Loop
		}

		r, l := utf8.DecodeRune(self.content[self.index:])
		if r == utf8.RuneError {
			err = errors.New("Invalid utf8 encoding")
			return
		}
		if r == '\'' {
			inchar = !inchar
		}
		if self.index == start {
			switch r {
			case '\n':
				tokType = newline
				self.index++
				break Loop
			case ':':
				tokType = begindef
				self.index++
				break Loop
			case ';':
				tokType = enddef
				self.index++
				break Loop
			case '{':
				incode++
				tokType = code
			case '|':
				tokType = alternate
				self.index++
				break Loop
			case '%':
				tokType = hfield
			case '\'':
				tokType = literal
			default:
				if unicode.IsUpper(r) {
					tokType = nonterm
				} else {
					tokType = term
				}
			}
		} else if incode > 0 && r == '{' {
			incode++
		} else if incode > 0 && r == '}' {
			incode--
		}
		if incode == 0 && !inchar && unicode.IsSpace(r) {
			break
		}
		self.index += l
	}
	word.tokType = TokType(tokType)
	word.text = string(self.content[start:self.index])
	if word.text == "%%" {
		word.tokType = TokType(separate)
	}
	return
}
