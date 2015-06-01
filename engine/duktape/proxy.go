package duktape

import "C"
import (
	"errors"
	"reflect"
	"strings"
	"unsafe"
)

var (
	UnexpectedPointer = errors.New("unexpected pointer")
	UndefinedProperty = errors.New("undefined property")
)

type proxy struct {
	structs map[unsafe.Pointer]interface{}
}

func newProxy(ctx *Context) *proxy {
	return &proxy{
		structs: make(map[unsafe.Pointer]interface{}, 0),
	}
}

func (p *proxy) register(s interface{}) unsafe.Pointer {
	ptr := C.malloc(1)
	p.structs[ptr] = s

	return ptr
}

func (p *proxy) retrieve(ptr unsafe.Pointer) interface{} {
	return p.structs[ptr]
}

func (p *proxy) has(t interface{}, k string) bool {
	_, err := p.getProperty(t, k)
	return err != UndefinedProperty
}

func (p *proxy) get(t interface{}, k string, recv interface{}) (interface{}, error) {
	f, err := p.getProperty(t, k)
	if err != nil {
		return nil, err
	}

	return f.Interface(), nil
}

func (p *proxy) set(t interface{}, k string, v, recv interface{}) (bool, error) {
	f, err := p.getProperty(t, k)
	if err != nil {
		return false, err
	}

	v = castNumberToGoType(f.Kind(), v)
	if !f.CanSet() {
		return false, nil
	}

	f.Set(reflect.ValueOf(v))
	return true, nil
}

func (p *proxy) enumerate(t interface{}) (interface{}, error) {
	return p.getPropertyNames(t)
}

func (p *proxy) getPropertyNames(t interface{}) ([]string, error) {
	v := reflect.ValueOf(t)
	names := make([]string, 0)
	switch v.Kind() {
	case reflect.Struct:
		cFields := v.NumField()
		for i := 0; i < cFields; i++ {
			names = append(names, v.Type().Field(i).Name)
		}
	}

	return names, nil
}

func (p *proxy) getProperty(t interface{}, key string) (reflect.Value, error) {
	v := reflect.ValueOf(t)
	r, found := p.getValueFromKind(key, v)
	if !found {
		return r, UndefinedProperty
	}

	return r, nil
}

func (p *proxy) getValueFromKind(key string, v reflect.Value) (reflect.Value, bool) {
	var value reflect.Value
	var found bool
	switch v.Kind() {
	case reflect.Ptr:
		value, found = p.getValueFromKindPtr(key, v)
	case reflect.Struct:
		value, found = p.getValueFromKindStruct(key, v)
	case reflect.Map:
		value, found = p.getValueFromKindMap(key, v)
	}

	if !found {
		return p.getMethod(key, v)
	}

	return value, found
}

func (p *proxy) getValueFromKindPtr(key string, v reflect.Value) (reflect.Value, bool) {
	r, found := p.getMethod(key, v)
	if !found {
		return p.getValueFromKind(key, v.Elem())
	}

	return r, found
}

func (p *proxy) getValueFromKindStruct(key string, v reflect.Value) (reflect.Value, bool) {
	key = strings.Title(key)
	r := v.FieldByName(key)

	return r, r.IsValid()
}

func (p *proxy) getValueFromKindMap(key string, v reflect.Value) (reflect.Value, bool) {
	kValue := reflect.ValueOf(key)
	r := v.MapIndex(kValue)

	return r, r.IsValid()
}

func (p *proxy) getMethod(key string, v reflect.Value) (reflect.Value, bool) {
	key = strings.Title(key)
	r := v.MethodByName(key)

	return r, r.IsValid()
}
