package parser

import (
	"log"
)

const Debug = false

func parserLog(format string, v ...interface{}) {
	if Debug {
		log.Printf(format, v...)
	}
}
