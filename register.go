package tba

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/olebedev/go-duktape"
)

func RegisterObject(ctx *duktape.Context, name string, o interface{}) error {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)

	bindings := make([]string, 0)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := getMethodName(name, method.Name)
		err := registerFunc(ctx, methodName, v.Method(i).Interface())
		if err != nil {
			return err
		}

		bindings = append(bindings, fmt.Sprintf(
			"%s: %s", strings.ToLower(method.Name), methodName,
		))
	}

	object := fmt.Sprintf("%s = { %s }", name, strings.Join(bindings, ", "))
	ctx.EvalString(object)

	return nil
}

func getMethodName(structName, methodName string) string {
	return fmt.Sprintf("%s__%s", structName, strings.ToLower(methodName))
}

func RegisterFunc(ctx *duktape.Context, f interface{}) error {
	return registerFunc(ctx, getFunctionName(f), f)
}

func registerFunc(ctx *duktape.Context, name string, f interface{}) error {
	return ctx.PushGoFunc(name, func(ctx *duktape.Context) int {
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
	case reflect.String:
		ctx.PushString(out[0].String())
	}
}
