package parser

import (
	"os"
	"testing"
)

func TestLLParser(t *testing.T) {
	in, _ := os.Open("/Users/Zach/go/src/github.com/Zach41/parser/input.y")
	out, _ := os.Create("yy.output.go")

	LLParser(in, out)

	in.Close()
	out.Close()
}
