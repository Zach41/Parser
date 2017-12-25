package parser

type Production struct {
	name string
	body []string
	code string
}

var MINTOKEN int
var MAXTOKEN int
var literalSet map[string]int
var tokenSet map[string]int
var symbolSet map[string]int

var prods []Production

var firsts map[int]int
var follows map[int]int

// header info variables
var packagename string
var modules []string
var defaultcode string
var unionTypes map[string]string
var termTypes map[string]string
var nontermTypes map[string]string
