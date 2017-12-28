package parser

import (
	"log"
)

const Debug = true

func parserLog(format string, v ...interface{}) {
	if Debug {
		log.Printf(format, v...)
	}
}

func type2Str(ttype TokType) string {
	switch ttype {
	case emptyTok:
		return "empty token"
	case nonterm:
		return "nonterm"
	case newline:
		return "newline"
	case begindef:
		return "begindef"
	case enddef:
		return "enddef"
	case alternate:
		return "alternate"
	case code:
		return "code"
	case hfield:
		return "header field"
	case separate:
		return "separate"
	case other:
		return "other token"
	}
	return ""
}
