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

	scanner := &Scanner{content: []byte(content), index: 0}
	ParserHeaders(scanner)

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

func checkMap(expected map[string]string, checked map[string]string, t *testing.T) {
	for vname, vtype := range expected {
		v, b := expected[vname]
		if !b || vtype != v {
			t.Errorf("Map checking error, expeted: {%s: %s}", vname, vtype)
		}
	}
}
