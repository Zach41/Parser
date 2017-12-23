package parser

type Production struct {
	name string
	body []string
	code string
}

var strtbl map[string]int
var firsts map[int]int
var follows map[int]int
var prods map[int][]Production

// header info variables
var packagename string
var modules []string
var defaultcode string
var unionTypes map[string]string
var termTypes map[string]string
var nontermTypes map[string]string
