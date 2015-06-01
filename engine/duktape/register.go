package duktape

import (
	"reflect"
	"strings"
	"unsafe"

	goduktape "github.com/olebedev/go-duktape"
)

const goProxyPtrProp = "\xff" + "goProxyPtrProp"

type Context struct {
	storage *storage
	*goduktape.Context
}

func NewContext() *Context {
	ctx := &Context{Context: goduktape.New()}
	ctx.storage = newStorage()

	return ctx
}

func (ctx *Context) SetRequireFunction(f interface{}) int {
	ctx.PushGlobalObject()
	ctx.GetPropString(-1, "Duktape")
	idx := ctx.PushGoFunction(f)
	ctx.PutPropString(-2, "modSearch")
	ctx.Pop()

	return idx
}

func (ctx *Context) ProxyGlobalInterface(name string, s interface{}) int {
	ctx.PushGlobalObject()
	obj := ctx.ProxyInterface(s)
	ctx.PutPropString(-2, name)
	ctx.Pop()

	return obj
}

func (ctx *Context) ProxyInterface(s interface{}) int {
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
		ctx.PushGoFunction(v.Method(i).Interface())
		ctx.PutPropString(obj, lowerCapital(t.Method(i).Name))
	}
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ctx.PushInt(int(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ctx.PushInt(int(v.Uint()))
	case reflect.Float64:
		ctx.PushNumber(v.Float())
	case reflect.String:
		ctx.PushString(v.String())
	case reflect.Struct:
		ctx.ProxyInterface(v.Interface())
	case reflect.Func:
		ctx.PushGoFunction(v.Interface())
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct {
			ctx.ProxyInterface(v.Interface())
			return nil
		}

		return ctx.PushValue(v.Elem())
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

func (ctx *Context) wrapFunction(f interface{}) func(ctx *goduktape.Context) int {
	tbaContext := ctx
	return func(ctx *goduktape.Context) int {
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

	value = castNumberToGoType(t.Kind(), value)
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
		value = ctx.GetProxy(index)

		if value == nil {
			if ctx.IsArray(index) {
				value = ctx.RequireSlice(index)
			} else {
				value = ctx.RequireMap(index)
			}
		}

	case goduktape.TypePointer:
		value = ctx.GetPointer(-1)
	case goduktape.TypeNull, goduktape.TypeUndefined, goduktape.TypeNone:
		value = nil
	default:
		value = "undefined"
	}

	return value
}

func (ctx *Context) GetProxy(index int) interface{} {
	ptr := ctx.getProxyPtrProp(index)
	if ptr == nil {
		return nil
	}

	return ctx.storage.get(ptr)
}

func (ctx *Context) getProxyPtrProp(index int) unsafe.Pointer {
	defer ctx.Pop()
	ctx.GetPropString(index, goProxyPtrProp)
	if !ctx.IsPointer(-1) {
		return nil
	}

	return ctx.GetPointer(-1)
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
	m := make(map[string]interface{}, 0)

	if !ctx.IsObject(index) {
		return m
	}

	ctx.Enum(index, goduktape.EnumOwnPropertiesOnly|goduktape.NoProxyBehavior)
	for ctx.IsObject(-1) {
		if !ctx.Next(-1, true) {
			break
		}

		m[ctx.RequireString(-2)] = ctx.RequireInterface(-1)
		ctx.Pop2()
	}

	ctx.Pop()
	return m
}

func (ctx *Context) callFunction(f interface{}, args []reflect.Value) int {
	var err error
	out := reflect.ValueOf(f).Call(args)
	out, err = ctx.handleReturnError(out)

	if err != nil {
		ctx.PushGoError(err)
		return goduktape.ErrRetError
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
		ctx.PushGoError(err)
		return goduktape.ErrRetInternal
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

func (ctx *Context) PushGoError(err error) {
	//fmt.Println(err)
	//ctx.Error(102, "foo %s", "qux")
}

func lowerCapital(name string) string {
	return strings.ToLower(name[:1]) + name[1:]
}

func castNumberToGoType(k reflect.Kind, v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch k {
	case reflect.Int:
		v = int(v.(float64))
	case reflect.Int8:
		v = int8(v.(float64))
	case reflect.Int16:
		v = int16(v.(float64))
	case reflect.Int32:
		v = int32(v.(float64))
	case reflect.Int64:
		v = int64(v.(float64))
	case reflect.Uint:
		v = uint(v.(float64))
	case reflect.Uint8:
		v = uint8(v.(float64))
	case reflect.Uint16:
		v = uint16(v.(float64))
	case reflect.Uint32:
		v = uint32(v.(float64))
	case reflect.Uint64:
		v = uint64(v.(float64))
	case reflect.Float32:
		v = float32(v.(float64))
	}

	return v
}
