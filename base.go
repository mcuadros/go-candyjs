package candyjs

import (
	"encoding/json"
	"reflect"
	"strings"
	"unsafe"

	"github.com/olebedev/go-duktape"
)

const goProxyPtrProp = "\xff" + "goProxyPtrProp"

type Context struct {
	storage *storage
	*duktape.Context
}

func NewContext() *Context {
	ctx := &Context{Context: duktape.New()}
	ctx.storage = newStorage()
	ctx.pushGlobalCandyJSObject()

	return ctx
}

func (ctx *Context) pushGlobalCandyJSObject() {
	ctx.PushGlobalObject()
	ctx.PushObject()
	ctx.PushObject()
	ctx.PutPropString(-2, "_functions")
	ctx.PushGoFunction(func(pckgName string) error {
		return ctx.pushPackage(pckgName)
	})
	ctx.PutPropString(-2, "require")
	ctx.PutPropString(-2, "CandyJS")
	ctx.Pop()

	ctx.EvalString(`CandyJS._call = function(ptr, args) {
		return CandyJS._functions[ptr].apply(null, args)
	}`)

	ctx.EvalString(`CandyJS.proxy = function(func) {
		ptr = Duktape.Pointer(func);
		CandyJS._functions[ptr] = func;

		return ptr;
	}`)
}

func (ctx *Context) SetRequireFunction(f interface{}) int {
	ctx.PushGlobalObject()
	ctx.GetPropString(-1, "Duktape")
	idx := ctx.PushGoFunction(f)
	ctx.PutPropString(-2, "modSearch")
	ctx.Pop()

	return idx
}

func (ctx *Context) PushGlobalType(name string, s interface{}) int {
	ctx.PushGlobalObject()
	cons := ctx.PushType(s)
	ctx.PutPropString(-2, name)
	ctx.Pop()

	return cons
}

func (ctx *Context) PushType(s interface{}) int {
	return ctx.PushGoFunction(func() {
		value := reflect.New(reflect.TypeOf(s))
		ctx.PushProxy(value.Interface())
	})
}

func (ctx *Context) PushGlobalProxy(name string, s interface{}) int {
	ctx.PushGlobalObject()
	obj := ctx.PushProxy(s)
	ctx.PutPropString(-2, name)
	ctx.Pop()

	return obj
}

func (ctx *Context) PushProxy(s interface{}) int {
	ptr := ctx.storage.add(s)

	obj := ctx.PushObject()
	ctx.PushPointer(ptr)
	ctx.PutPropString(-2, goProxyPtrProp)

	ctx.PushGlobalObject()
	ctx.GetPropString(-1, "Proxy")
	ctx.Dup(obj)

	ctx.PushObject()
	ctx.PushGoFunction(p.enumerate)
	ctx.PutPropString(-2, "enumerate")
	ctx.PushGoFunction(p.enumerate)
	ctx.PutPropString(-2, "ownKeys")
	ctx.PushGoFunction(p.get)
	ctx.PutPropString(-2, "get")
	ctx.PushGoFunction(p.set)
	ctx.PutPropString(-2, "set")
	ctx.PushGoFunction(p.has)
	ctx.PutPropString(-2, "has")
	ctx.New(2)

	ctx.Remove(-2)
	ctx.Remove(-2)

	ctx.PushPointer(ptr)
	ctx.PutPropString(-2, goProxyPtrProp)

	return obj
}

func (ctx *Context) PushGlobalStruct(name string, s interface{}) (int, error) {
	ctx.PushGlobalObject()
	obj, err := ctx.PushStruct(s)
	if err != nil {
		return -1, err
	}

	ctx.PutPropString(-2, name)
	ctx.Pop()

	return obj, nil
}

func (ctx *Context) PushStruct(s interface{}) (int, error) {
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)

	obj := ctx.PushObject()
	ctx.pushStructMethods(obj, t, v)

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	return obj, ctx.pushStructFields(obj, t, v)
}

func (ctx *Context) pushStructFields(obj int, t reflect.Type, v reflect.Value) error {
	fCount := t.NumField()
	for i := 0; i < fCount; i++ {
		value := v.Field(i)

		if value.Kind() != reflect.Ptr || !value.IsNil() {
			fieldName := lowerCapital(t.Field(i).Name)
			if fieldName == t.Field(i).Name {
				continue
			}

			if err := ctx.PushValue(value); err != nil {
				return err
			}

			ctx.PutPropString(obj, fieldName)
		}
	}

	return nil
}

func (ctx *Context) pushStructMethods(obj int, t reflect.Type, v reflect.Value) {
	mCount := t.NumMethod()
	for i := 0; i < mCount; i++ {
		methodName := lowerCapital(t.Method(i).Name)

		if methodName == t.Method(i).Name {
			continue
		}

		ctx.PushGoFunction(v.Method(i).Interface())
		ctx.PutPropString(obj, methodName)

	}
}

func (ctx *Context) PushGlobalInterface(name string, v interface{}) error {
	return ctx.PushGlobalValue(name, reflect.ValueOf(v))
}

func (ctx *Context) PushInterface(v interface{}) error {
	return ctx.PushValue(reflect.ValueOf(v))
}

func (ctx *Context) PushGlobalValue(name string, v reflect.Value) error {
	ctx.PushGlobalObject()
	if err := ctx.PushValue(v); err != nil {
		return err
	}

	ctx.PutPropString(-2, name)
	ctx.Pop()

	return nil
}

func (ctx *Context) PushValue(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Interface:
		return ctx.PushValue(v.Elem())
	case reflect.Bool:
		ctx.PushBoolean(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		ctx.PushInt(int(v.Int()))
	case reflect.Int64: //Caveat: lose of precession casting to float64
		ctx.PushNumber(float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		ctx.PushUint(uint(v.Uint()))
	case reflect.Uint64: //Caveat: lose of precession casting to float64
		ctx.PushNumber(float64(v.Uint()))
	case reflect.Float64:
		ctx.PushNumber(v.Float())
	case reflect.String:
		ctx.PushString(v.String())
	case reflect.Struct:
		ctx.PushProxy(v.Interface())
	case reflect.Func:
		ctx.PushGoFunction(v.Interface())
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct {
			ctx.PushProxy(v.Interface())
			return nil
		}

		return ctx.PushValue(v.Elem())

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			ctx.PushString(string(v.Interface().([]byte)))
		} else {
			ctx.PushNull()
		}
	default:
		//Returns nul if the Kind is not supported
		ctx.PushNull()
	}

	return nil
}

func (ctx *Context) PushGlobalValues(name string, vs []reflect.Value) error {
	ctx.PushGlobalObject()
	if err := ctx.PushValues(vs); err != nil {
		return err
	}

	ctx.PutPropString(-2, name)
	ctx.Pop()

	return nil
}

func (ctx *Context) PushValues(vs []reflect.Value) error {
	arr := ctx.PushArray()
	for i, v := range vs {
		if err := ctx.PushValue(v); err != nil {
			return err
		}

		ctx.PutPropIndex(arr, uint(i))
	}

	return nil
}

func (ctx *Context) PushGlobalGoFunction(name string, f interface{}) (int, error) {
	return ctx.Context.PushGlobalGoFunction(name, ctx.wrapFunction(f))
}

func (ctx *Context) PushGoFunction(f interface{}) int {
	return ctx.Context.PushGoFunction(ctx.wrapFunction(f))
}

func (ctx *Context) wrapFunction(f interface{}) func(ctx *duktape.Context) int {
	tbaContext := ctx
	return func(ctx *duktape.Context) int {
		args := tbaContext.getFunctionArgs(f)
		return tbaContext.callFunction(f, args)
	}
}

func (ctx *Context) getFunctionArgs(f interface{}) []reflect.Value {
	def := reflect.ValueOf(f).Type()
	isVariadic := def.IsVariadic()
	inCount := def.NumIn()

	top := ctx.GetTopIndex()
	args := make([]reflect.Value, 0)
	for index := 0; index <= top; index++ {
		var t reflect.Type
		if (index+1) < inCount || (index < inCount && !isVariadic) {
			t = def.In(index)
		} else if isVariadic {
			t = def.In(inCount - 1).Elem()
		}

		args = append(args, ctx.getValueFromContext(index, t))
	}

	//Optional args
	argc := len(args)
	if inCount > argc {
		for i := argc; i < inCount; i++ {
			//Avoid send empty slice when variadic
			if isVariadic && i-1 < inCount {
				break
			}

			args = append(args, reflect.Zero(def.In(i)))
		}
	}

	return args
}

func (ctx *Context) getValueFromContext(index int, t reflect.Type) reflect.Value {
	if proxy := ctx.GetProxy(index); proxy != nil {
		return reflect.ValueOf(proxy)
	}

	if ctx.IsPointer(index) {
		return ctx.getFunction(index, t)
	}

	return ctx.getValueUsingJson(index, t)
}

func (ctx *Context) GetProxy(index int) interface{} {
	if !ctx.IsObject(index) {
		return nil
	}

	ptr := ctx.getProxyPtrProp(index)
	if ptr == nil {
		return nil
	}

	return ctx.storage.get(ptr)
}

func (ctx *Context) getFunction(index int, t reflect.Type) reflect.Value {
	ptr := ctx.GetPointer(index)

	return reflect.MakeFunc(t, ctx.wrapDuktapePointer(ptr, t))
}

func (ctx *Context) wrapDuktapePointer(
	ptr unsafe.Pointer,
	t reflect.Type,
) func(in []reflect.Value) []reflect.Value {
	return func(in []reflect.Value) []reflect.Value {
		ctx.PushGlobalObject()
		ctx.GetPropString(-1, "CandyJS")
		obj := ctx.NormalizeIndex(-1)
		ctx.PushString("_call")
		ctx.PushPointer(ptr)
		ctx.PushValues(in)
		ctx.CallProp(obj, 2)

		return ctx.getCallResult(t)
	}
}

func (ctx *Context) getCallResult(t reflect.Type) []reflect.Value {
	result := make([]reflect.Value, 0)

	oCount := t.NumOut()
	if oCount == 1 {
		result = append(result, ctx.getValueFromContext(-1, t.Out(0)))
		return result
	}

	return result
}

func (ctx *Context) getProxyPtrProp(index int) unsafe.Pointer {
	defer ctx.Pop()
	ctx.GetPropString(index, goProxyPtrProp)
	if !ctx.IsPointer(-1) {
		return nil
	}

	return ctx.GetPointer(-1)
}

func (ctx *Context) getValueUsingJson(index int, t reflect.Type) reflect.Value {
	v := reflect.New(t).Interface()

	js := ctx.JsonEncode(index)
	if len(js) == 0 {
		return reflect.Zero(t)
	}

	err := json.Unmarshal([]byte(js), v)
	if err != nil {
		panic(err)
	}

	return reflect.ValueOf(v).Elem()
}

func (ctx *Context) callFunction(f interface{}, args []reflect.Value) int {
	var err error
	out := reflect.ValueOf(f).Call(args)
	out, err = ctx.handleReturnError(out)

	if err != nil {
		return duktape.ErrRetError
	}

	if len(out) == 0 {
		return 1
	}

	if len(out) > 1 {
		err = ctx.PushValues(out)
	} else {
		err = ctx.PushValue(out[0])
	}

	if err != nil {
		return duktape.ErrRetInternal
	}

	return 1
}

func (ctx *Context) handleReturnError(out []reflect.Value) ([]reflect.Value, error) {
	c := len(out)
	if c == 0 {
		return out, nil
	}

	last := out[c-1]
	if last.Type().Name() == "error" {
		if !last.IsNil() {
			return nil, last.Interface().(error)
		}

		return out[:c-1], nil
	}

	return out, nil
}

func lowerCapital(name string) string {
	return strings.ToLower(name[:1]) + name[1:]
}
