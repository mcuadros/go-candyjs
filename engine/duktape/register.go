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
	for index := 1; index <= top; index++ {
		i := index - 1
		var t reflect.Type
		if index < inCount || (index == inCount && !isVariadic) {
			t = def.In(i)
		} else if isVariadic {
			t = def.In(inCount - 1).Elem()
		}

		args = append(args, ctx.getValueFromContext(index, t))
	}

	//Optional args
	argc := len(args)
	if inCount > argc {
		for i := argc; i < inCount; i++ {
			args = append(args, reflect.Zero(def.In(i)))
		}
	}

	return args
}

func (ctx *Context) getValueFromContext(index int, t reflect.Type) reflect.Value {
	value := ctx.RequireInterface(index)
	if value == nil {
		return reflect.Zero(t)
	}

	switch t.Kind() {
	case reflect.Int:
		value = int(value.(float64))
	case reflect.Int8:
		value = int8(value.(float64))
	case reflect.Int16:
		value = int16(value.(float64))
	case reflect.Int32:
		value = int32(value.(float64))
	case reflect.Int64:
		value = int64(value.(float64))
	case reflect.Uint:
		value = uint(value.(float64))
	case reflect.Uint8:
		value = uint8(value.(float64))
	case reflect.Uint16:
		value = uint16(value.(float64))
	case reflect.Uint32:
		value = uint32(value.(float64))
	case reflect.Uint64:
		value = uint64(value.(float64))
	case reflect.Float32:
		value = float32(value.(float64))
	}

	return reflect.ValueOf(value)
}

func (ctx *Context) RequireInterface(index int) interface{} {
	var value interface{}

	switch ctx.GetType(index) {
	case goduktape.TypeString:
		value = ctx.RequireString(index)
	case goduktape.TypeNumber:
		value = ctx.RequireNumber(index)
	case goduktape.TypeBoolean:
		value = ctx.RequireBoolean(index)
	case goduktape.TypeObject:
		if ctx.IsArray(index) {
			value = ctx.RequireSlice(index)
		} else {
			value = ctx.RequireMap(index)
		}
	case goduktape.TypeNull, goduktape.TypeUndefined, goduktape.TypeNone:
		value = nil
	default:
		value = "undefined"
	}

	return value
}

func (ctx *Context) RequireSlice(index int) []interface{} {
	s := make([]interface{}, 0)
	var i uint
	for ctx.GetPropIndex(index, i) {
		i++
		s = append(s, ctx.RequireInterface(-1))
	}

	return s
}

func (ctx *Context) RequireMap(index int) map[string]interface{} {
	ctx.Enum(index, goduktape.EnumOwnPropertiesOnly)

	m := make(map[string]interface{}, 0)
	for ctx.Next(-1, true) {
		m[ctx.RequireString(-2)] = ctx.RequireInterface(-1)
		ctx.Pop2()
	}

	return m
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
