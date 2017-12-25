package parser

import (
	"os"
	"strings"
)

func ParseHeaders(scanner *Scanner) {
	for err, word := scanner.NextWord(); err == nil; err, word = scanner.NextWord() {
		if word.tokType == separate {
			break
		}
		if word.tokType == newline {
			continue
		}
		if word.tokType == hfield {
			switch word.text {
			case "%package":
				err, name := scanner.NextWord()
				if err != nil {
					parserLog("Parse `package` error: %s", err.Error())
					os.Exit(1)
				}
				packagename = name.text
			case "%defaultcode":
				err, tcode := scanner.NextWord()
				if err != nil || tcode.tokType != code {
					parserLog("Parse `defaultcode` error: %s", err.Error())
					os.Exit(1)
				}
				defaultcode = tcode.text
			case "%import":
				parseModules(scanner)
			case "%union":
				parseUnionTypes(scanner)
			default:
				if strings.Index(word.text, "%token") == 0 {
					parseSymbolTypes(termTypes, word.text[7:len(word.text)-1], scanner)
				}
				if strings.Index(word.text, "%type") == 0 {
					parseSymbolTypes(nontermTypes, word.text[6:len(word.text)-1], scanner)
				}
			}
		}
	}
}

func ParseGrammars(scanner *Scanner) {
	for err, word := scanner.NextWord(); err == nil && word.tokType != separate; err, word = scanner.NextWord() {
		if word.tokType == newline {
			continue
		}
		eatSymbol(&word)
		production := Production{name: word.text}
		parserLog("Parsing Grammar: %s", word.text)

		startGrammar := true
		for err, word = scanner.NextWord(); err == nil && word.tokType != enddef; err, word = scanner.NextWord() {
			if startGrammar && word.tokType != begindef {
				parserLog("Expected grammar begin def in grammar %s", production.name)
				os.Exit(1)
			}
			if !startGrammar && word.tokType != alternate {
				parserLog("Expected alternate after first rule in grammar %s", production.name)
				os.Exit(1)
			}
			if startGrammar {
				startGrammar = !startGrammar
			}

			body, bodyCode := parseGrammarBody(scanner)
			production.body = body
			production.code = bodyCode
			prods = append(prods, production)
		}
	}
}

func MergeSymbols(literals map[string]int, tokens map[string]int, symbols map[string]int) map[string]int {
	merged := make(map[string]int)
	merged[""] = 0
	merged["$"] = 1
	MINTOKEN, MAXTOKEN = 0, 0
	for lit, id := range literals {
		merged[lit] = id + 2
		if MINTOKEN < id+2 {
			MINTOKEN = id + 2
		}
	}
	MINTOKEN += 1
	for tok, id := range tokens {
		merged[tok] = MINTOKEN + id
		if MAXTOKEN < MINTOKEN+id {
			MAXTOKEN = MINTOKEN + id
		}
	}
	for sym, id := range symbols {
		merged[sym] = MAXTOKEN + id + 1
	}
	return merged
}

func parseModules(scanner *Scanner) {
	for err, word := scanner.NextWord(); err == nil && word.tokType != newline; err, word = scanner.NextWord() {
		if word.tokType == newline {
			continue
		}
		moduleName := word.text
		modules = append(modules, moduleName)
	}
}

func parseUnionTypes(scanner *Scanner) {
	err, text := scanner.NextWord()
	if err != nil || text.tokType != code {
		parserLog("Parse `union` error: %s", err.Error())
		os.Exit(1)
	}
	code_text := strings.Trim(text.text, " \n")
	code_text = code_text[1 : len(code_text)-1]
	parserLog("----------Union:\n%s\n", code_text)
	codeScanner := Scanner{content: []byte(code_text), index: 0}
	for err, word := codeScanner.NextWord(); err == nil; err, word = codeScanner.NextWord() {
		if word.tokType == newline {
			continue
		}
		vType := ""
		for err, typeTok := codeScanner.NextWord(); err == nil && typeTok.tokType != newline; err, typeTok = codeScanner.NextWord() {
			vType += typeTok.text
		}
		tname := word.text
		unionTypes[tname] = vType
	}
}

func parseSymbolTypes(symTbl map[string]string, typeName string, scanner *Scanner) {
	for err, word := scanner.NextWord(); err == nil && word.tokType != newline; err, word = scanner.NextWord() {
		parserLog("Symbol %s", word.text)
		symName := word.text
		symTbl[symName] = typeName
	}
}

func parseGrammarBody(scanner *Scanner) ([]string, string) {
	body := make([]string, 0)
	bodyCode := ""
Loop:
	for {
		err, word := scanner.NextWord()
		if err != nil {
			parserLog("Parsing grammar body error: %s", err.Error())
			os.Exit(1)
		}
		switch word.tokType {
		case nonterm:
			fallthrough
		case term:
			fallthrough
		case literal:
			eatSymbol(&word)
			body = append(body, word.text)
		case code:
			bodyCode = word.text
		default:
			break Loop
		}
	}
	parserLog("Body: %v", body)
	return body, bodyCode
}

func eatSymbol(word *WordTok) {
	parserLog("Eating {%s, %s}", word.text, type2Str(word.tokType))
	switch word.tokType {
	case literal:
		if _, b := literalSet[word.text]; !b {
			literalSet[word.text] = len(literalSet)
		}
	case nonterm:
		if _, b := symbolSet[word.text]; !b {
			symbolSet[word.text] = len(symbolSet)
		}
	case term:
		if _, b := tokenSet[word.text]; !b {
			tokenSet[word.text] = len(tokenSet)
		}
	}
}
