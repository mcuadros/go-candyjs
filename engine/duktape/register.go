package duktape

import (
	"fmt"
	"reflect"
	"strings"

	goduktape "github.com/olebedev/go-duktape"
)

type Context struct {
	*goduktape.Context
}

func NewContext() *Context {
	return &Context{goduktape.NewContext()}
}

func (ctx *Context) RegisterInstance(name string, o interface{}) error {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)

	bindings := make([]string, 0)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := getMethodName(name, method.Name)
		err := ctx.RegisterFunc(methodName, v.Method(i).Interface())
		if err != nil {
			return err
		}

		bindings = append(bindings, fmt.Sprintf(
			"%s: %s", lowerCapital(method.Name), methodName,
		))
	}

	object := fmt.Sprintf("%s = { %s }", name, strings.Join(bindings, ", "))
	ctx.EvalString(object)

	return nil
}

func (ctx *Context) RegisterFunc(name string, f interface{}) error {
	tbaContext := ctx
	return ctx.PushGoFunc(name, func(ctx *goduktape.Context) int {
		args := tbaContext.getFunctionArgs(f)
		tbaContext.callFunction(f, args)

		return 1
	})
}

func (ctx *Context) getFunctionArgs(f interface{}) []reflect.Value {
	def := reflect.ValueOf(f).Type()
	isVariadic := def.IsVariadic()
	inCount := def.NumIn()

	top := ctx.GetTopIndex()
	args := make([]reflect.Value, 0)
	for i := 1; i <= top; i++ {
		var k reflect.Kind
		if i < inCount || (i == inCount && !isVariadic) {
			k = def.In(i - 1).Kind()
		} else if isVariadic {
			k = def.In(inCount - 1).Elem().Kind()
		}

		args = append(args, ctx.getValueFromContext(i, k))
	}

	return args
}

func (ctx *Context) getValueFromContext(index int, k reflect.Kind) reflect.Value {
	var value interface{}
	switch { //The order is important
	case ctx.IsString(index):
		value = ctx.RequireString(index)
	case ctx.IsNumber(index):
		return ctx.getNumberFromContext(index, k)
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

func (ctx *Context) getNumberFromContext(index int, k reflect.Kind) reflect.Value {
	var value interface{}
	switch k {
	case reflect.Int:
		value = int(ctx.RequireNumber(index))
	case reflect.Float64:
		value = float64(ctx.RequireNumber(index))
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
}

func (ctx *Context) pushValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Int:
		ctx.PushInt(int(v.Int()))
	case reflect.Float64:
		ctx.PushNumber(v.Float())
	case reflect.String:
		ctx.PushString(v.String())
	}
}

func getMethodName(structName, methodName string) string {
	return fmt.Sprintf("%s__%s", structName, lowerCapital(methodName))
}

func lowerCapital(name string) string {
	return strings.ToLower(name[:1]) + name[1:]
}
