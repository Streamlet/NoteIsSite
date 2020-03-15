package util

import "fmt"

func Assert(expr bool, format string, a ...interface{}) {
	if !expr {
		panic(fmt.Sprintf(format, a...))
	}
}
