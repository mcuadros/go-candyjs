package runtime

import (
	"github.com/mcuadros/tba/engine/duktape"
)

var ctx *duktape.Context

func init() {
	ctx = duktape.NewContext()
	ctx.PushGlobalGoFunction("include", include)
	ctx.PushGlobalGoFunction("require", require)
	ctx.PushGlobalStruct("console", NewConsole())
	ctx.PushGlobalStruct("os", NewOS())
}

func GetContext() *duktape.Context {
	return ctx
}
