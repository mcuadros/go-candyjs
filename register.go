package tba

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/olebedev/go-duktape"
)

type Context struct {
	*duktape.Context
}

func NewContext() *Context {
	return &Context{duktape.NewContext()}
}

func (ctx *Context) RegisterInstance(name string, o interface{}) error {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)

	bindings := make([]string, 0)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := getMethodName(name, method.Name)
		err := ctx.registerFunc(methodName, v.Method(i).Interface())
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

func (ctx *Context) RegisterFunc(f interface{}) error {
	return ctx.registerFunc(getFunctionName(f), f)
}

func (ctx *Context) registerFunc(name string, f interface{}) error {
	tbaContext := ctx
	return ctx.PushGoFunc(name, func(ctx *duktape.Context) int {
		args := tbaContext.getFunctionArgs()
		tbaContext.callFunction(f, args)

		return 1
	})
}

func (ctx *Context) getFunctionArgs() []reflect.Value {
	top := ctx.GetTopIndex()
	args := make([]reflect.Value, 0)
	for i := 1; i <= top; i++ {
		args = append(args, ctx.getValueFromContext(i))
	}

	return args
}

func (ctx *Context) getValueFromContext(index int) reflect.Value {
	var value interface{}
	switch { //The order is important
	case ctx.IsString(index):
		value = ctx.RequireString(index)
	case ctx.IsNumber(index):
		value = int(ctx.RequireNumber(index))
	case ctx.IsBoolean(index):
		value = ctx.RequireBoolean(index)
	case ctx.IsNull(index), ctx.IsNan(index), ctx.IsUndefined(index):
		value = nil
	case ctx.IsArray(index):
		value = "array"
	case ctx.IsObject(index):
		value = "object"
	default:
		value = "undefined"
	}

	return reflect.ValueOf(value)
}

func (ctx *Context) callFunction(f interface{}, args []reflect.Value) {
	out := reflect.ValueOf(f).Call(args)
	out = ctx.handleReturnError(out)

	if len(out) == 0 {
		return
	}

	if len(out) > 1 {
		ctx.pushValues(out)
	} else {
		ctx.pushValue(out[0])
	}
}

func (ctx *Context) handleReturnError(out []reflect.Value) []reflect.Value {
	c := len(out)
	if c == 0 {
		return out
	}

	last := out[c-1]
	if last.Type().Name() == "error" {
		if !last.IsNil() {
			fmt.Println(last.Interface())
		}

		return out[:c-1]
	}

	return out
}

func (ctx *Context) pushValues(vs []reflect.Value) {
	arr := ctx.PushArray()
	for i, v := range vs {
		ctx.pushValue(v)
		ctx.PutPropIndex(arr, uint(i))
	}

	fmt.Println(vs)
	ctx.Pop()
}

func (ctx *Context) pushValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Int:
		ctx.PushInt(int(v.Int()))
	case reflect.String:
		ctx.PushString(v.String())
	}
}

func getMethodName(structName, methodName string) string {
	return fmt.Sprintf("%s__%s", structName, strings.ToLower(methodName))
}

func getFunctionName(i interface{}) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	return strings.Split(fn, ".")[1]
}
