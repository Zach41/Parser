package parser

import (
	"testing"
)

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

func checkWord(scanner *Scanner, t *testing.T, ttype TokType, text string) {
	err, word := scanner.NextWord()
	if err != nil || word.tokType != ttype || word.text != text {
		t.Errorf("Expeted {%s %s}, got {%s %s}.",
			text, type2Str(ttype), word.text, type2Str(word.tokType))
	}
}

func TestScannerBasic(t *testing.T) {
	content := `%package main # Test

%import scanner fmt

%defaultcode {
    fmt.println("Test")
}

%% 

Calc : Add
     ;

Mult : '*' Mult
     | '/' Mult
     ;

Add : floating
    | integer
    ;`
	scanner := Scanner{content: []byte(content), index: 0}
	checkWord(&scanner, t, hfield, "%package")
	checkWord(&scanner, t, term, "main")
	// checkWord(&scanner, t, term, "#")
	// checkWord(&scanner, t, term, "Test")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, hfield, "%import")
	checkWord(&scanner, t, term, "scanner")
	checkWord(&scanner, t, term, "fmt")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, hfield, "%defaultcode")
	checkWord(&scanner, t, code, "{\n    fmt.println(\"Test\")\n}")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, separate, "%%")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, nonterm, "Calc")
	checkWord(&scanner, t, begindef, ":")
	checkWord(&scanner, t, nonterm, "Add")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, enddef, ";")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, nonterm, "Mult")
	checkWord(&scanner, t, begindef, ":")
	checkWord(&scanner, t, term, "'*'")
	checkWord(&scanner, t, nonterm, "Mult")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, alternate, "|")
	checkWord(&scanner, t, term, "'/'")
	checkWord(&scanner, t, nonterm, "Mult")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, enddef, ";")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, nonterm, "Add")
	checkWord(&scanner, t, begindef, ":")
	checkWord(&scanner, t, term, "floating")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, alternate, "|")
	checkWord(&scanner, t, term, "integer")
	checkWord(&scanner, t, newline, "\n")
	checkWord(&scanner, t, enddef, ";")
	err, _ := scanner.NextWord()

	if err.Error() != "End of File" {
		t.Errorf("Expetected End of File")
	}
}
