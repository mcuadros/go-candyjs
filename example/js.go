package main

import (
	"fmt"

	"github.com/mcuadros/tba"
	"github.com/olebedev/go-duktape"
)

func main() {
	ctx := duktape.NewContext()
	tba.RegisterFunc(ctx, multiply)
	tba.RegisterFunc(ctx, printf)

	ctx.EvalFile("js.js")
}

func multiply(a, b int) int {
	return a * b
}

func printf(format string, str ...interface{}) {
	fmt.Printf(format, str...)
}
