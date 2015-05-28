package runtime

import (
	"fmt"

	"github.com/mcuadros/tba/engine/duktape"
)

var ctx *duktape.Context

func init() {
	ctx = duktape.NewContext()
	ctx.PushGlobalGoFunction("global", nil)
	ctx.PushGlobalGoFunction("module", nil)
	ctx.PushGlobalGoFunction("require", require)
	ctx.PushGlobalGoFunction("include", include)
	ctx.PushGlobalStruct("console", NewConsole())
	ctx.PushGlobalStruct("os", NewOS())
}

func GetContext() *duktape.Context {
	return ctx
}

type Values struct {
	Foo int
}
