package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func LLParser(in *os.File, out *os.File) {
	content, err := ioutil.ReadAll(in)
	if err != nil {
		fmt.Printf("Reading content err: %s\n", err.Error())
		return
	}

	scanner := &Scanner{content: content, index: 0}

	// init values
	literalSet = make(map[string]int)
	tokenSet = make(map[string]int)
	symbolSet = make(map[string]int)
	prods = make([]Production, 0)
	MINTOKEN, MAXTOKEN = 0, 0
	modules = make([]string, 0)
	unionTypes = make(map[string]string)
	termTypes = make(map[string]string)
	nontermTypes = make(map[string]string)

	ParseHeaders(scanner)
	ParseGrammars(scanner)

	restCode := scanner.Reminder()

	mergedSymbols := MergeSymbols(literalSet, tokenSet, symbolSet)
	firsts := ComputeFirsts(prods, mergedSymbols, MAXTOKEN)
	follows := ComputeFollows(prods, mergedSymbols, firsts)
	lltable := ComputeLLTable(prods, mergedSymbols, firsts, follows, MAXTOKEN+1, len(mergedSymbols)-1)

	printFile(lltable, mergedSymbols, out)
	out.Write(restCode)
}

func printFile(lltable map[int][]int,
	tokens map[string]int,
	out *os.File) {
	out.WriteString("// A LL Grammar Parser, writen by Zach41\n// Version 0.1\n\n")

	// package name
	out.WriteString(fmt.Sprintf("package %s\n\n", packagename))
	// modules
	out.WriteString("import (\n")
	out.WriteString("\t\"fmt\"\n")
	for _, module := range modules {
		if module == "fmt" {
			continue
		}
		out.WriteString(fmt.Sprintf("\t\"%s\"\n", module))
	}
	out.WriteString(")\n\n")

	out.WriteString("const (\n")
	out.WriteString(fmt.Sprintf("\tMAXTOKEN = %d\n", MAXTOKEN))
	out.WriteString(fmt.Sprintf("\tMINTOKEN = %d\n", MINTOKEN))
	out.WriteString(")\n\n")
	// write yytype
	out.WriteString("type yytype struct {\n")
	for tname, ttype := range unionTypes {
		out.WriteString(fmt.Sprintf("\t%s    %s\n", tname, ttype))
	}
	out.WriteString("}\n\n")

	// Stack is a helper struct
	out.WriteString(`type Stack struct {
    values [2048]interface{}    // stack size is limited to 2048
    top    int
}

func (stack *Stack) pop() interface{} {
    if stack.top < 0 {
        return nil
    }
    ret := stack.values[stack.top]
    stack.top -= 1
    return ret
}

func (stack *Stack) push(value interface{}) {
    if stack.top >= 2047 {
        return
    } else {
        stack.top += 1
        stack.values[stack.top] = value
    }
}

func (stack *Stack) empty() bool {
    return stack.top < 0
}

func NewStack() *Stack {
    stack := &Stack{top: -1}
    return stack
}

func word2Idx(word string) int {
    wordIdx := 1
    if idx, b := yycharmap[word]; b {
        wordIdx = idx
    } else {
        litword := fmt.Sprintf("'%s'", word)
        if idx, b := yycharmap[litword]; b {
            wordIdx = idx
        } else {
            fmt.Printf("Unrecognized token: %s\n", word)
            wordIdx = -1
        }
    }
    return wordIdx
}

`)
	out.WriteString("func bodyOfIdx(idx int) []int {\n")
	out.WriteString("\tbodyIdxes := make([]int, 0)\n")
	out.WriteString("\tswitch idx {\n")
	for idx, prod := range prods {
		out.WriteString(fmt.Sprintf("\tcase %d:\n", idx))
		for _, body := range prod.body {
			out.WriteString(fmt.Sprintf("\t\tbodyIdxes = append(bodyIdxes, yycharmap[\"%s\"])\n", body))
		}
	}
	out.WriteString("}\n")
	out.WriteString("\treturn bodyIdxes\n}\n\n")

	// write yytable
	out.WriteString("var yytable = map[int][]int{\n")
	for k, row := range lltable {
		out.WriteString(fmt.Sprintf("\t%d: []int{ ", k))
		for i, v := range row {
			if i == len(row)-1 {
				out.WriteString(fmt.Sprintf("%d },\n", v))
			} else {
				out.WriteString(fmt.Sprintf("%d, ", v))
			}
		}
	}
	out.WriteString("}\n\n")

	// write all symbol mappings
	out.WriteString("var yycharmap = map[string]int{\n")
	for name, idx := range tokens {
		if len(name) == 0 {
			out.WriteString(fmt.Sprintf("\t\"\": %d,\n", idx))
		} else {
			out.WriteString(fmt.Sprintf("\t\"%s\": %d,\n", name, idx))
		}
	}
	out.WriteString("}\n\n")

	// running code when reduction happends
	// idx: which production is reducing, start with 0
	// values: current values stack
	// return a yytype value
	out.WriteString("func yyruncode(idx int, values *Stack) *yytype {\n")
	out.WriteString("\tlhs := &yytype{}\n")
	out.WriteString("\tswitch idx {\n")
	for i, prod := range prods {
		var codeStr string
		if len(prod.code) == 0 {
			codeStr = defaultcode
		} else {
			codeStr = prod.code
		}
		parserLog("Original Code:\n%s", codeStr)
		out.WriteString(fmt.Sprintf("\tcase %d:\n", i))
		// if len(prod.code) == 0 {
		// 	out.WriteString("\t\treturn lhs\n")
		// } else {
		lhsValue := fmt.Sprintf("\t\tlhs.%s", nontermTypes[prod.name])
		prodCode := strings.Replace(codeStr, "$$", lhsValue, -1)
		rhsIdx := 1
		for _, rhsName := range prod.body {
			if tokens[rhsName] < MINTOKEN {
				rhsIdx += 1
				continue
			}
			// out.WriteString("\t\trhs_%d := values.pop().(*yytype)\n")
			rhsVar := fmt.Sprintf("rhs_%d", rhsIdx)
			out.WriteString(fmt.Sprintf("\t\t%s := values.pop().(*yytype)\n", rhsVar))
			var rhsValue string
			if tokens[rhsName] <= MAXTOKEN {
				rhsValue = fmt.Sprintf("%s.%s", rhsVar, termTypes[rhsName])
			} else {
				rhsValue = fmt.Sprintf("%s.%s", rhsVar, nontermTypes[rhsName])
			}
			oldStr := fmt.Sprintf("$%d", rhsIdx)
			parserLog("Replacing %s to %s in:\n%s", oldStr, rhsValue, prodCode)
			prodCode = strings.Replace(prodCode, oldStr, rhsValue, -1)
			rhsIdx += 1
		}
		out.WriteString(fmt.Sprintf("\t\t%s\n", prodCode))
		out.WriteString("\t\tvalues.push(lhs)\n\t\treturn lhs\n")
		// }
	}
	out.WriteString("\t}\n\t return lhs\n}\n\n")

	out.WriteString(`func yyparser(nextWord func()(bool, string, *yytype)) *yytype {
    values := NewStack()
    prods := NewStack()
    stack := NewStack()

    eof, word, yyval := nextWord()
    if eof {
        word = "$"
    }
    wordIdx := word2Idx(word)
    if wordIdx < 0 {
        return nil
    }
    stack.push(MAXTOKEN + 1)

    for !stack.empty() {
        top := stack.pop().(int)
        if top > MAXTOKEN {
            if yytable[top][wordIdx] == -1 {
                fmt.Println("Error Parsing")
                return nil
            }
            prods.push(top)
            bodyIdxes := bodyOfIdx(top)
            for i := len(bodyIdxes); i>=0; i-- {
                 stack.push(bodyIdxes[i])
            }
        } else {
            if top != wordIdx {
                fmt.Println("Not Matching Terminal Or Literal")
                return nil
            }
            if top >= MINTOKEN {
                values.push(yyval)
            }
            eof, word, yyval = nextWord()
            if eof {
                word = "$"
            }
            wordIdx = word2Idx(word)
        }
    }

    return yyexecution(values, prods)
}

func yyexecution(values *Stack, prods *Stack) *yytype {
    var yyval *yytype
    for !prods.empty() {
        prod := prods.pop().(int)
        yyval = yyruncode(prod, values)
    }
    return yyval
}

`)

}
