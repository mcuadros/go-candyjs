package runtime

import (
	"github.com/mcuadros/tba/engine/duktape"
)

var ctx *duktape.Context

func init() {
	ctx = duktape.NewContext()
	ctx.RegisterFunc("include", include)
	ctx.RegisterFunc("require", require)

	ctx.RegisterFunc("mf", func(a, b float64) float64 {
		return a * b
	})

	ctx.RegisterInstance("console", NewConsole())
}

func GetContext() *duktape.Context {
	return ctx
}
