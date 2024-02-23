package utils

import "fmt"

func Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
