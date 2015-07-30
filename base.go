package candyjs

import (
	"encoding/json"
	"reflect"
	"unsafe"

	"github.com/olebedev/go-duktape"
)

const goProxyPtrProp = "\xff" + "goProxyPtrProp"

// Context represents a Duktape thread and its call and value stacks.
type Context struct {
	storage *storage
	mutex   *mutex
	Duktape *duktape.Context
}

// NewContext returns a new Context
func NewContext() *Context {
	ctx := &Context{Duktape: duktape.New()}
	ctx.storage = newStorage()
	ctx.mutex = new(mutex)
	ctx.pushGlobalCandyJSObject()

	return ctx
}

func (ctx *Context) NewTransaction() *Transaction {
	return newTransaction(ctx)
}

func (ctx *Context) pushGlobalCandyJSObject() {
	ctx.Duktape.PushGlobalObject()
	ctx.Duktape.PushObject()
	ctx.Duktape.PushObject()
	ctx.Duktape.PutPropString(-2, "_functions")
	ctx.PushGoFunction(noTransaction, func(pckgName string) error {
		return ctx.pushPackage(pckgName)
	})
	ctx.Duktape.PutPropString(-2, "require")
	ctx.Duktape.PutPropString(-2, "CandyJS")
	ctx.Duktape.Pop()

	ctx.Duktape.EvalString(`CandyJS._call = function(ptr, args) {
		return CandyJS._functions[ptr].apply(null, args)
	}`)

	ctx.Duktape.EvalString(`CandyJS.proxy = function(func) {
		ptr = Duktape.Pointer(func);
		CandyJS._functions[ptr] = func;

		return ptr;
	}`)
}

// SetRequireFunction sets the modSearch function into the Duktape JS object
// http://duktape.org/guide.html#builtin-duktape-modsearch-modloade
func (ctx *Context) SetRequireFunction(f interface{}) int {
	ctx.Duktape.PushGlobalObject()
	ctx.Duktape.GetPropString(-1, "Duktape")
	idx := ctx.PushGoFunction(noTransaction, f)
	ctx.Duktape.PutPropString(-2, "modSearch")
	ctx.Duktape.Pop()

	return idx
}

// PushGlobalType like PushType but pushed to the global object
func (ctx *Context) PushGlobalType(t transactionID, name string, s interface{}) int {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	ctx.Duktape.PushGlobalObject()
	cons := ctx.PushType(t, s)
	ctx.Duktape.PutPropString(-2, name)
	ctx.Duktape.Pop()

	return cons
}

// PushType push a constructor for the type of the given value, this constructor
// returns an empty instance of the type. The value passed is discarded, only
// is used for retrieve the time, instead of require pass a `reflect.Type`.
func (ctx *Context) PushType(t transactionID, s interface{}) int {
	return ctx.PushGoFunction(t, func() {
		value := reflect.New(reflect.TypeOf(s))
		ctx.PushProxy(noTransaction, value.Interface())
	})
}

// PushGlobalProxy like PushProxy but pushed to the global object
func (ctx *Context) PushGlobalProxy(t transactionID, name string, v interface{}) int {
	ctx.Duktape.PushGlobalObject()
	obj := ctx.PushProxy(t, v)
	ctx.Duktape.PutPropString(-2, name)
	ctx.Duktape.Pop()

	return obj
}

// PushProxy push a proxified pointer of the given value to the stack, this
// refence will be stored on an internal storage. The pushed objects has
// the exact same methods and properties from the original value.
// http://duktape.org/guide.html#virtualization-proxy-object
func (ctx *Context) PushProxy(t transactionID, v interface{}) int {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	ptr := ctx.storage.add(v)

	obj := ctx.Duktape.PushObject()
	ctx.Duktape.PushPointer(ptr)
	ctx.Duktape.PutPropString(-2, goProxyPtrProp)

	ctx.Duktape.PushGlobalObject()
	ctx.Duktape.GetPropString(-1, "Proxy")
	ctx.Duktape.Dup(obj)

	ctx.Duktape.PushObject()
	ctx.PushGoFunction(t, p.enumerate)
	ctx.Duktape.PutPropString(-2, "enumerate")
	ctx.PushGoFunction(t, p.enumerate)
	ctx.Duktape.PutPropString(-2, "ownKeys")
	ctx.PushGoFunction(t, p.get)
	ctx.Duktape.PutPropString(-2, "get")
	ctx.PushGoFunction(t, p.set)
	ctx.Duktape.PutPropString(-2, "set")
	ctx.PushGoFunction(t, p.has)
	ctx.Duktape.PutPropString(-2, "has")
	ctx.Duktape.New(2)

	ctx.Duktape.Remove(-2)
	ctx.Duktape.Remove(-2)

	ctx.Duktape.PushPointer(ptr)
	ctx.Duktape.PutPropString(-2, goProxyPtrProp)

	return obj
}

// PushGlobalStruct like PushStruct but pushed to the global object
func (ctx *Context) PushGlobalStruct(t transactionID, name string, s interface{}) (int, error) {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	ctx.Duktape.PushGlobalObject()
	obj, err := ctx.PushStruct(t, s)
	if err != nil {
		return -1, err
	}

	ctx.Duktape.PutPropString(-2, name)
	ctx.Duktape.Pop()

	return obj, nil
}

// PushStruct push a object to the stack with the same methods and properties
// the pushed object is a copy, any change made on JS is not reflected on the
// Go instance.
func (ctx *Context) PushStruct(t transactionID, s interface{}) (int, error) {
	ts := reflect.TypeOf(s)
	vs := reflect.ValueOf(s)

	obj := ctx.Duktape.PushObject()
	ctx.pushStructMethods(t, obj, ts, vs)

	if ts.Kind() == reflect.Ptr {
		vs = vs.Elem()
		ts = vs.Type()
	}

	return obj, ctx.pushStructFields(t, obj, ts, vs)
}

func (ctx *Context) pushStructFields(t transactionID, obj int, ts reflect.Type, vs reflect.Value) error {
	fCount := ts.NumField()
	for i := 0; i < fCount; i++ {
		value := vs.Field(i)

		if value.Kind() != reflect.Ptr || !value.IsNil() {
			fieldName := ts.Field(i).Name
			if !isExported(fieldName) {
				continue
			}

			if err := ctx.pushValue(t, value); err != nil {
				return err
			}

			ctx.Duktape.PutPropString(obj, nameToJavaScript(fieldName))
		}
	}

	return nil
}

func (ctx *Context) pushStructMethods(t transactionID, obj int, ts reflect.Type, vs reflect.Value) {
	mCount := ts.NumMethod()
	for i := 0; i < mCount; i++ {
		methodName := ts.Method(i).Name
		if !isExported(methodName) {
			continue
		}

		ctx.PushGoFunction(t, vs.Method(i).Interface())
		ctx.Duktape.PutPropString(obj, nameToJavaScript(methodName))

	}
}

// PushGlobalInterface like PushInterface but pushed to the global object
func (ctx *Context) PushGlobalInterface(t transactionID, name string, v interface{}) error {
	return ctx.pushGlobalValue(t, name, reflect.ValueOf(v))
}

// PushInterface push any type of value to the stack, the following types are
// supported:
//  - Bool
//  - Int, Int8, Int16, Int32, Uint, Uint8, Uint16, Uint32 and Uint64
//  - Float32 and Float64
//  - Strings and []byte
//  - Structs
//  - Functions with any signature
//
// Please read carefully the following notes:
//  - The pointers are resolved and the value is pushed
//  - Structs are pushed ussing PushProxy, if you want to make a copy use PushStruct
//  - Int64 and UInt64 are supported but before push it to the stack are casted
//    to float64
//  - Any unsuported value is pushed as a null
func (ctx *Context) PushInterface(t transactionID, v interface{}) error {
	return ctx.pushValue(t, reflect.ValueOf(v))
}

func (ctx *Context) pushGlobalValue(t transactionID, name string, v reflect.Value) error {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	ctx.Duktape.PushGlobalObject()
	if err := ctx.pushValue(t, v); err != nil {
		return err
	}

	ctx.Duktape.PutPropString(-2, name)
	ctx.Duktape.Pop()

	return nil
}

func (ctx *Context) pushValue(t transactionID, v reflect.Value) error {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	if !v.IsValid() {
		ctx.Duktape.PushNull()
		return nil
	}

	switch v.Kind() {
	case reflect.Interface:
		return ctx.pushValue(t, v.Elem())
	case reflect.Bool:
		ctx.Duktape.PushBoolean(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		ctx.Duktape.PushInt(int(v.Int()))
	case reflect.Int64: //Caveat: lose of precession casting to float64
		ctx.Duktape.PushNumber(float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		ctx.Duktape.PushUint(uint(v.Uint()))
	case reflect.Uint64: //Caveat: lose of precession casting to float64
		ctx.Duktape.PushNumber(float64(v.Uint()))
	case reflect.Float64:
		ctx.Duktape.PushNumber(v.Float())
	case reflect.String:
		ctx.Duktape.PushString(v.String())
	case reflect.Struct:
		ctx.PushProxy(t, v.Interface())
	case reflect.Func:
		ctx.PushGoFunction(t, v.Interface())
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct {
			ctx.PushProxy(t, v.Interface())
			return nil
		}

		return ctx.pushValue(t, v.Elem())
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			ctx.Duktape.PushString(string(v.Interface().([]byte)))
			return nil
		}

		fallthrough
	default:
		js, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}

		ctx.Duktape.PushString(string(js))
		ctx.Duktape.JsonDecode(-1)
	}

	return nil
}

func (ctx *Context) pushGlobalValues(t transactionID, name string, vs []reflect.Value) error {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	ctx.Duktape.PushGlobalObject()
	if err := ctx.pushValues(t, vs); err != nil {
		return err
	}

	ctx.Duktape.PutPropString(-2, name)
	ctx.Duktape.Pop()

	return nil
}

func (ctx *Context) pushValues(t transactionID, vs []reflect.Value) error {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	arr := ctx.Duktape.PushArray()
	for i, v := range vs {
		if err := ctx.pushValue(t, v); err != nil {
			return err
		}

		ctx.Duktape.PutPropIndex(arr, uint(i))
	}

	return nil
}

// PushGlobalGoFunction like PushGoFunction but pushed to the global object
func (ctx *Context) PushGlobalGoFunction(t transactionID, name string, f interface{}) (int, error) {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	return ctx.Duktape.PushGlobalGoFunction(name, ctx.wrapFunction(f))
}

// PushGoFunction push a native Go function of any signature to the stack.
// A pointer to the function is stored in the internals of the context and
// collected by the duktape GC removing any reference in Go also.
//
// The most common types are supported as input arguments, also the variadic
// functions can be used.
//
// You can use JS functions as arguments but you should wrapper it with the
// helper `CandyJS.proxy`. Example:
// 	ctx.PushGlobalGoFunction("test", func(fn func(int, int) int) {
//		...
//	})
//
//	ctx.PevalString(`test(CandyJS.proxy(function(a, b) { return a * b; }));`)
//
// The structs can be delivered to the functions in three ways:
//  - In-line representation as plain JS objects: `{'int':42}`
//  - Using a previous pushed type using `PushGlobalType`: `new MyModel`
//  - Using a previous pushed instance using `PushGlobalProxy`
//
// All other types are loaded into Go using `json.Unmarshal` internally
//
// The following types are not supported chans, complex64 or complex128, and
// the types rune, byte and arrays are not tested.
//
// The returns are handled in the following ways:
//  - The result of functions with a single return value like `func() int` is
//    pushed directly to the stack.
//  - Functions with a n return values like `func() (int, int)` are pushed as
//    an array. The errors are removed from this array.
//  - Returns of functions with a trailling error like `func() (string, err)`:
//    if err is not nil an error is throw in the context, and the other values
//    are discarded. IF err is nil, the values are pushed to the stack, following
//    the previuos rules.
//
// All the non erros returning values are pushed following the same rules of
// `PushInterface` method
func (ctx *Context) PushGoFunction(t transactionID, f interface{}) int {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	return ctx.Duktape.PushGoFunction(ctx.wrapFunction(f))
}

func (ctx *Context) wrapFunction(f interface{}) func(ctx *duktape.Context) int {
	return func(god *duktape.Context) int {
		args := ctx.getFunctionArgs(f)
		return ctx.callFunction(NoTransaction, f, args)
	}
}

func (ctx *Context) getFunctionArgs(f interface{}) []reflect.Value {
	def := reflect.ValueOf(f).Type()
	isVariadic := def.IsVariadic()
	inCount := def.NumIn()

	top := ctx.Duktape.GetTopIndex()

	var args []reflect.Value
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
	if proxy := ctx.getProxy(index); proxy != nil {
		return reflect.ValueOf(proxy)
	}

	if ctx.Duktape.IsPointer(index) {
		return ctx.getFunction(index, t)
	}

	return ctx.getValueUsingJSON(index, t)
}

func (ctx *Context) getProxy(index int) interface{} {
	if !ctx.Duktape.IsObject(index) {
		return nil
	}

	ptr := ctx.getProxyPtrProp(index)
	if ptr == nil {
		return nil
	}

	return ctx.storage.get(ptr)
}

func (ctx *Context) getFunction(index int, t reflect.Type) reflect.Value {
	ptr := ctx.Duktape.GetPointer(index)

	return reflect.MakeFunc(t, ctx.wrapDuktapePointer(ptr, t))
}

func (ctx *Context) wrapDuktapePointer(
	ptr unsafe.Pointer,
	t reflect.Type,
) func(in []reflect.Value) []reflect.Value {
	return func(in []reflect.Value) []reflect.Value {
		ctx.Duktape.PushGlobalObject()
		ctx.Duktape.GetPropString(-1, "CandyJS")
		obj := ctx.Duktape.NormalizeIndex(-1)
		ctx.Duktape.PushString("_call")
		ctx.Duktape.PushPointer(ptr)
		ctx.pushValues(noTransaction, in)
		ctx.Duktape.CallProp(obj, 2)

		return ctx.getCallResult(t)
	}
}

func (ctx *Context) getCallResult(t reflect.Type) []reflect.Value {
	var result []reflect.Value

	oCount := t.NumOut()
	if oCount == 1 {
		result = append(result, ctx.getValueFromContext(-1, t.Out(0)))
	} else if oCount > 1 {
		if ctx.Duktape.GetLength(-1) != oCount {
			panic("Invalid count of return value on proxied function.")
		}

		idx := ctx.Duktape.NormalizeIndex(-1)
		for i := 0; i < oCount; i++ {
			ctx.Duktape.GetPropIndex(idx, uint(i))
			result = append(result, ctx.getValueFromContext(-1, t.Out(i)))
		}
	}

	return result
}

func (ctx *Context) getProxyPtrProp(index int) unsafe.Pointer {
	defer ctx.Duktape.Pop()
	ctx.Duktape.GetPropString(index, goProxyPtrProp)
	if !ctx.Duktape.IsPointer(-1) {
		return nil
	}

	return ctx.Duktape.GetPointer(-1)
}

func (ctx *Context) getValueUsingJSON(index int, t reflect.Type) reflect.Value {
	v := reflect.New(t).Interface()

	js := ctx.Duktape.JsonEncode(index)
	if len(js) == 0 {
		return reflect.Zero(t)
	}

	err := json.Unmarshal([]byte(js), v)
	if err != nil {
		panic(err)
	}

	return reflect.ValueOf(v).Elem()
}

func (ctx *Context) callFunction(t transactionID, f interface{}, args []reflect.Value) int {
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
		err = ctx.pushValues(t, out)
	} else {
		err = ctx.pushValue(t, out[0])
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

//EvalString evaluate the Ecmascript source code.
func (ctx *Context) EvalString(t transactionID, src string) error {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	return ctx.Duktape.PevalString(src)
}

//EvalFile evaluate the Ecmascript contained in the given file.
func (ctx *Context) EvalFile(t transactionID, path string) error {
	ctx.mutex.Lock(t)
	defer ctx.mutex.Unlock(t)

	return ctx.Duktape.PevalFile(path)
}
