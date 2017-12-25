package parser

import (
	"testing"
)

func TestParseHeaders(t *testing.T) {
	content := `%package main     # Set the package of the generated file to "main"

%import scanner fmt os strconv  # import (
                                #   scanner
                                #   fmt
                                #   os
                                #   strconv
                                # )

# Replace the default { $$ = $1 } rule code with this custom code.
%defaultcode {
    fmt.Println("Default code. Assigning", $1, " to ", $$, "."); $$ = $1
}

# Define the custom value type for tokens
%union {
    fval float
    ival int
    op func(float)float
}

# Associate the "floating" terminal with the type of fval float
%token<fval> floating

# Associate the "integer" terminal with the type of ival int
%token<ival> integer

# Associate the "Calc", "Num", "Mult", and "Add" nonterminals with the type of fval float 
%type<fval> Calc Num Mult Add

# Associate the "MultA" and "AddA" nonterminals with the type of op func(float)float
%type<op> MultA AddA`
	unionTypes = make(map[string]string)
	termTypes = make(map[string]string)
	nontermTypes = make(map[string]string)
	modules = make([]string, 0)

	scanner := &Scanner{content: []byte(content), index: 0}
	ParseHeaders(scanner)

	if packagename != "main" {
		t.Errorf("Expected package name: main, Got: %s ", packagename)
	}
	expectedModules := []string{"scanner", "fmt", "os", "strconv"}
	for i, text := range modules {
		if expectedModules[i] != text {
			t.Errorf("Expected package: %s, Got: %s", expectedModules[i], text)
		}
	}
	expectedCode := `{
    fmt.Println("Default code. Assigning", $1, " to ", $$, "."); $$ = $1
}`
	if defaultcode != expectedCode {
		t.Errorf("Expected default code:\n%s\n, Got:\n%s\n", expectedCode, defaultcode)
	}

	expectedUnions := map[string]string{"fval": "float", "ival": "int", "op": "func(float)float"}
	checkMap(expectedUnions, unionTypes, t)

	expectedTermTypes := map[string]string{
		"floating": "fval",
		"integer":  "ival",
	}
	checkMap(expectedTermTypes, termTypes, t)

	expectedNontermTypes := map[string]string{
		"Calc":  "fval",
		"Num":   "fval",
		"Mult":  "fval",
		"Add":   "fval",
		"MultA": "op",
		"AddA":  "op",
	}
	checkMap(expectedNontermTypes, nontermTypes, t)
}

func TestParseGrammar(t *testing.T) {
	content := `Calc : Add        # This will use the code in %defaultcode
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
	scanner := &Scanner{content: []byte(content), index: 0}
	literalSet = make(map[string]int)
	tokenSet = make(map[string]int)
	symbolSet = make(map[string]int)
	prods = make([]Production, 0)

	ParseGrammars(scanner)

	expectedLiterals := map[string]int{
		"'*'": 0,
		"'/'": 1,
		"'+'": 2,
		"'-'": 3,
	}
	expectedTokens := map[string]int{
		"floating": 0,
		"integer":  1,
	}
	expectedSymbols := map[string]int{
		"Calc":  0,
		"Add":   1,
		"Mult":  2,
		"Num":   3,
		"MultA": 4,
		"AddA":  5,
	}
	expectedProd1 := "Calc: [Add]"
	expectedProd2 := `Mult: [Num, MultA] { fmt.Println("1. Found Mult->'*' Num."); $$ = $2($1) }`
	expectedProd3 := `MultA: ['*', Mult] { fmt.Println("2. Found MultA->'*' Mult."); $$ = mult($2) }`
	expectedProd4 := `MultA: ['/', Mult] { fmt.Println("3. Found MultA->'/' Mult."); $$ = div($2) }`
	expectedProd5 := `MultA: [] { fmt.Println("4. Found MultA->{}."); $$ = noop }`
	expectedProd6 := `Add: [Mult, AddA] { fmt.Println("5. Found Add->Mult AddA."); $$ = $2($1) }`
	expectedProd7 := `AddA: ['+', Add] { fmt.Println("6. Found AddA->'+' Add"); $$ = plus($2) }`
	expectedProd8 := `AddA: ['-', Add] { fmt.Println("7. Found AddA->'-' Add"); $$ = minus($2) }`
	expectedProd9 := `AddA: [] { fmt.Println("8. Found AddA->{}"); $$ = noop}`
	expectedProd10 := `Num: [floating] { fmt.Println("9. Found Num->floating. Forwarding value", $1); $$ = float($1) }`
	expectedProd11 := `Num: [integer] { fmt.Println("10. Found Num->integer. Forwarding value", $1); $$ = float($1) }`
	expectedProds := []string{
		expectedProd1, expectedProd2, expectedProd3,
		expectedProd4, expectedProd5, expectedProd6,
		expectedProd7, expectedProd8, expectedProd9,
		expectedProd10, expectedProd11,
	}

	checkMap2(expectedLiterals, literalSet, t)
	checkMap2(expectedTokens, tokenSet, t)
	checkMap2(expectedSymbols, symbolSet, t)

	if len(prods) != len(expectedProds) {
		t.Errorf("Production Parsing Error")
	}
	for i, prod := range prods {
		gotProd := prod2string(&prod)
		parserLog("Got Production: %s", gotProd)
		parserLog("Expected Production: %s", expectedProds[i])
		if gotProd != expectedProds[i] {
			t.Errorf("Production Parsing Error.\n\tExpected: %s\n\tGot: %s",
				expectedProds[i], gotProd)
		}
	}
	mergedSymbols := MergeSymbols(literalSet, tokenSet, symbolSet)
	expectedMergedSymbols := map[string]int{
		"":         0,
		"$":        1,
		"'*'":      2,
		"'/'":      3,
		"'+'":      4,
		"'-'":      5,
		"floating": 6,
		"integer":  7,
		"Calc":     8,
		"Add":      9,
		"Mult":     10,
		"Num":      11,
		"MultA":    12,
		"AddA":     13,
	}
	checkMap2(expectedMergedSymbols, mergedSymbols, t)
	if MINTOKEN != 6 {
		t.Errorf("Expected MINTOKEN to be 5")
	}
	if MAXTOKEN != 7 {
		t.Errorf("Expected MAXTOKEN to be 6")
	}
}

func checkMap(expected map[string]string, checked map[string]string, t *testing.T) {
	for vname, vtype := range expected {
		v, b := checked[vname]
		if !b || vtype != v {
			t.Errorf("Map checking error, expeted: {%s: %s}, got {%s: %s}", vname, vtype, vname, v)
		}
	}
}

func checkMap2(expected map[string]int, checked map[string]int, t *testing.T) {
	for vname, id := range expected {
		v, b := checked[vname]
		if !b || id != v {
			t.Errorf("Map checking error, expeted: {%s: %d} got {%s: %d}", vname, id, vname, v)
		}
	}
}

func prod2string(prod *Production) string {
	prodStr := prod.name

	prodStr += ": ["
	for i, v := range prod.body {
		prodStr += v
		if i != len(prod.body)-1 {
			prodStr += ", "
		}
	}
	prodStr += "]"
	if len(prod.code) > 0 {
		prodStr += " " + prod.code
	}
	return prodStr
}
