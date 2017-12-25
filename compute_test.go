package parser

import (
	"sort"
	"testing"
)

const content = `Calc : Add        # This will use the code in %defaultcode
     ;

Mult : Num MultA        { fmt.Println("1. Found Mult->'*' Num."); $$ = $2($1) }
     ;
    
MultA : '*' Mult        { fmt.Println("2. Found MultA->'*' Mult."); $$ = mult($2) }
      | '/' Mult        { fmt.Println("3. Found MultA->'/' Mult."); $$ = div($2) }
      |                 { fmt.Println("4. Found MultA->{}."); $$ = noop }
      ;

Add : Mult AddA         { fmt.Println("5. Found Add->Mult AddA."); $$ = $2($1) }
    ;

AddA : '+' Add          { fmt.Println("6. Found AddA->'+' Add"); $$ = plus($2) }
     | '-' Add          { fmt.Println("7. Found AddA->'-' Add"); $$ = minus($2) }
     |                  { fmt.Println("8. Found AddA->{}"); $$ = noop}
     ;

Num : floating          { fmt.Println("9. Found Num->floating. Forwarding value", $1); $$ = float($1) }
    | integer           { fmt.Println("10. Found Num->integer. Forwarding value", $1); $$ = float($1) }
    ;

%%`

func TestComputeFirsts(t *testing.T) {
	literalSet = make(map[string]int)
	tokenSet = make(map[string]int)
	symbolSet = make(map[string]int)
	prods = make([]Production, 0)

	scanner := &Scanner{content: []byte(content), index: 0}
	ParseGrammars(scanner)
	mergedSymbols := MergeSymbols(literalSet, tokenSet, symbolSet)

	parserLog("All Symbols: %v", mergedSymbols)

	firsts := ComputeFirsts(prods, mergedSymbols, MAXTOKEN)

	expectedFirsts := map[string][]int{
		"":         []int{0},
		"'*'":      []int{2},
		"'/'":      []int{3},
		"'+'":      []int{4},
		"'-'":      []int{5},
		"floating": []int{6},
		"integer":  []int{7},
		"Calc":     []int{6, 7},
		"Add":      []int{6, 7},
		"Mult":     []int{6, 7},
		"AddA":     []int{0, 4, 5},
		"MultA":    []int{0, 2, 3},
		"Num":      []int{6, 7},
	}
	parserLog("Got Firsts: %v", firsts)

	for k, rhs := range expectedFirsts {
		if _, b := firsts[k]; !b {
			t.Errorf("Expected key [%s] in Firsts", k)
		}
		lhs := firsts[k]
		if !cmpArraySorted(lhs, rhs) {
			t.Errorf("Expected %v for key %s, got %v", rhs, k, lhs)
		}
	}
}

func TestComputeFollows(t *testing.T) {
	literalSet = make(map[string]int)
	tokenSet = make(map[string]int)
	symbolSet = make(map[string]int)
	prods = make([]Production, 0)

	scanner := &Scanner{content: []byte(content), index: 0}
	ParseGrammars(scanner)

	mergedSymbols := MergeSymbols(literalSet, tokenSet, symbolSet)
	firsts := ComputeFirsts(prods, mergedSymbols, MAXTOKEN)
	follows := ComputeFollows(prods, mergedSymbols, firsts)

	expectedFollows := map[string][]int{
		"Calc":  []int{1},
		"Add":   []int{1},
		"AddA":  []int{1},
		"Mult":  []int{1, 4, 5},
		"MultA": []int{1, 4, 5},
		"Num":   []int{1, 2, 3, 4, 5},
	}
	parserLog("Got Follows:\n%v", follows)
	for k, rhs := range expectedFollows {
		if _, b := follows[k]; !b {
			t.Errorf("Expected key: %s in follows", k)
		}
		lhs := follows[k]
		if !cmpArraySorted(lhs, rhs) {
			t.Errorf("Expected %v for key %s, got %v", rhs, k, lhs)
		}
	}
}

func TestComputeLLTable(t *testing.T) {
	literalSet = make(map[string]int)
	tokenSet = make(map[string]int)
	symbolSet = make(map[string]int)
	prods = make([]Production, 0)

	scanner := &Scanner{content: []byte(content), index: 0}
	ParseGrammars(scanner)

	mergedSymbols := MergeSymbols(literalSet, tokenSet, symbolSet)
	firsts := ComputeFirsts(prods, mergedSymbols, MAXTOKEN)
	follows := ComputeFollows(prods, mergedSymbols, firsts)
	lltable := ComputeLLTable(prods, mergedSymbols, firsts, follows, MAXTOKEN+1, len(mergedSymbols)-1)

	expectedLLTable := map[int][]int{
		8:  []int{-1, -1, -1, -1, -1, -1, 0, 0},
		9:  []int{-1, -1, -1, -1, -1, -1, 5, 5},
		13: []int{-1, 8, -1, -1, 6, 7, -1, -1},
		10: []int{-1, -1, -1, -1, -1, -1, 1, 1},
		12: []int{-1, 4, 2, 3, 4, 4, -1, -1},
		11: []int{-1, -1, -1, -1, -1, -1, 9, 10},
	}
	parserLog("Got LL Table:\n%v\n", lltable)
	for k, v := range expectedLLTable {
		if _, b := lltable[k]; !b {
			t.Errorf("Expected a row for %d", k)
		}
		for i, pidx := range v {
			if lltable[k][i] != pidx {
				t.Errorf("\nExpected: %v\nGot: %v\n", v, lltable[k])
			}
		}
	}
}

func TestComputeFirstsFollows(t *testing.T) {
	content := `
E  : T E2
   ;

E2 : '+' T E2
   |
   ;

T  : F T2
   ;

T2 : '*' F T2
   |
   ;

F  : '(' E ')'
   | id
   ;`

	scanner := &Scanner{content: []byte(content), index: 0}
	literalSet = make(map[string]int)
	tokenSet = make(map[string]int)
	symbolSet = make(map[string]int)
	prods = make([]Production, 0)

	ParseGrammars(scanner)
	mergedSymbols := MergeSymbols(literalSet, tokenSet, symbolSet)
	firsts := ComputeFirsts(prods, mergedSymbols, MAXTOKEN)
	follows := ComputeFollows(prods, mergedSymbols, firsts)

	expectedMergedSymbols := map[string]int{
		"":    0,
		"$":   1,
		"'+'": 2,
		"'*'": 3,
		"'('": 4,
		"')'": 5,
		"id":  6,
		"E":   7,
		"T":   8,
		"E2":  9,
		"F":   10,
		"T2":  11,
	}
	checkMap2(expectedMergedSymbols, mergedSymbols, t)

	expectedFirsts := map[string][]int{
		"":    []int{0},
		"T":   []int{4, 6},
		"F":   []int{4, 6},
		"E":   []int{4, 6},
		"E2":  []int{2, 0},
		"T2":  []int{3, 0},
		"'+'": []int{2},
		"'*'": []int{3},
		"'('": []int{4},
		"')'": []int{5},
		"id":  []int{6},
	}

	for k, rhs := range expectedFirsts {
		if _, b := firsts[k]; !b {
			t.Errorf("Expected key [%s] in Firsts", k)
		}
		lhs := firsts[k]
		if !cmpArraySorted(lhs, rhs) {
			t.Errorf("Expected %v for key %s, got %v", rhs, k, lhs)
		}
	}

	expectedFollows := map[string][]int{
		"E":  []int{1, 5},
		"E2": []int{1, 5},
		"T":  []int{1, 2, 5},
		"T2": []int{1, 2, 5},
		"F":  []int{1, 2, 3, 5},
	}
	for k, rhs := range expectedFollows {
		if _, b := follows[k]; !b {
			t.Errorf("Expected key: %s in follows", k)
		}
		lhs := follows[k]
		if !cmpArraySorted(lhs, rhs) {
			t.Errorf("Expected %v for key %s, got %v", rhs, k, lhs)
		}
	}

	lltable := ComputeLLTable(prods, mergedSymbols, firsts, follows, MAXTOKEN+1, len(mergedSymbols)-1)
	expectedLLTable := map[int][]int{
		7:  []int{-1, -1, -1, -1, 0, -1, 0},
		9:  []int{-1, 2, 1, -1, -1, 2, -1},
		8:  []int{-1, -1, -1, -1, 3, -1, 3},
		11: []int{-1, 5, 5, 4, -1, 5, -1},
		10: []int{-1, -1, -1, -1, 6, -1, 7},
	}
	parserLog("Got LL Table:\n%v\n", lltable)
	for k, v := range expectedLLTable {
		if _, b := lltable[k]; !b {
			t.Errorf("Expected a row for %d", k)
		}
		for i, pidx := range v {
			if lltable[k][i] != pidx {
				t.Errorf("\nExpected: %v\nGot: %v\n", v, lltable[k])
			}
		}
	}
}

func cmpArraySorted(lhs []int, rhs []int) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	sort.Ints(lhs)
	sort.Ints(rhs)

	for idx, v := range lhs {
		if v != rhs[idx] {
			return false
		}
	}
	return true
}
