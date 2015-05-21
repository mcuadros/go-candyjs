package tba

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/olebedev/go-duktape"
)

func RegisterFunc(ctx *duktape.Context, f interface{}) {
	ctx.PushGoFunc(getFunctionName(f), func(ctx *duktape.Context) int {
		args := getFunctionArgs(ctx)
		callFunction(ctx, f, args)

		return 1
	})
}

func getFunctionName(i interface{}) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	return strings.Split(fn, ".")[1]
}

func getFunctionArgs(ctx *duktape.Context) []reflect.Value {
	top := ctx.GetTopIndex()
	args := make([]reflect.Value, 0)
	for i := 1; i <= top; i++ {
		args = append(args, getValueFromContext(ctx, i))
	}

	return args
}

func getValueFromContext(ctx *duktape.Context, index int) reflect.Value {
	var value interface{}
	switch {
	case ctx.IsString(index):
		value = ctx.RequireString(index)
	case ctx.IsNumber(index):
		value = int(ctx.RequireNumber(index))
	default:
		value = "undefined"
	}

	return reflect.ValueOf(value)
}

func callFunction(ctx *duktape.Context, f interface{}, args []reflect.Value) {
	out := reflect.ValueOf(f).Call(args)
	l := len(out)

	if l == 0 {
		return
	}

	if l > 1 {
		panic("function not allowed with multiple return values")
	}

	switch out[0].Kind() {
	case reflect.Int:
		ctx.PushInt(int(out[0].Int()))
	}
}
