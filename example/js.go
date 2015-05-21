package main

import (
	"fmt"

	"github.com/mcuadros/tba"
	"github.com/olebedev/go-duktape"
)

func main() {
	ctx := duktape.NewContext()
	tba.RegisterFunc(ctx, multiply)
	tba.RegisterFunc(ctx, sprintf)
	tba.RegisterObject(ctx, "console", &Log{})

	ctx.EvalFile("js.js")
}

func multiply(a, b int) int {
	return a * b
}

func sprintf(format string, str ...interface{}) string {
	return fmt.Sprintf(format, str...)
}

type Log struct{}

func (l *Log) Log(str ...interface{}) {
	fmt.Println(str...)
}
