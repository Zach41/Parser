package parser

import (
	"os"
	"strings"
)

func ParserHeaders(scanner *Scanner) {
	for err, word := scanner.NextWord(); err == nil; err, word = scanner.NextWord() {
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
	codeScanner := Scanner{content: []byte(code_text), index: 0}
	for err, word := codeScanner.NextWord(); err == nil; err, word = codeScanner.NextWord() {
		if word.tokType == newline {
			continue
		}
		// err, typeTok := codeScanner.NextWord()
		// if err != nil {
		// 	parserLog("Parse `union types` error: %s", err.Error())
		// 	os.Exit(1)
		// }
		vType := ""
		for err, typeTok := codeScanner.NextWord(); err == nil && word.tokType != newline; err, typeTok = codeScanner.NextWord() {
			vType += typeTok.text
		}
		tname := word.text
		unionTypes[tname] = vType
	}
}

func parseSymbolTypes(symTbl map[string]string, typeName string, scanner *Scanner) {
	for err, word := scanner.NextWord(); err == nil && word.tokType != newline; err, word = scanner.NextWord() {
		if word.tokType == newline {
			continue
		}
		symName := word.text
		symTbl[symName] = typeName
	}
}

func ParserGrammars(scanner *Scanner) {}
