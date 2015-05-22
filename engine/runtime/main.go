package runtime

import (
	"github.com/mcuadros/tba/engine/duktape"
)

var ctx *duktape.Context

func init() {
	ctx = duktape.NewContext()
	ctx.RegisterFunc("include", include)
	ctx.RegisterFunc("require", require)
	ctx.RegisterInstance("console", NewConsole())
}

func GetContext() *duktape.Context {
	return ctx
}
