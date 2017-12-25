package parser

import (
	"log"
)

const Debug = true

func parserLog(format string, v ...interface{}) {
	if Debug {
		log.Printf(format, v...)
	}
}
