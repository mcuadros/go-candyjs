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

func (p *proxy) has(t map[string]interface{}, k string) bool {
	_, err := p.getField(t, k)
	return err != UndefinedProperty
}

func (p *proxy) get(t map[string]interface{}, k string, recv interface{}) (interface{}, error) {
	f, err := p.getField(t, k)
	if err != nil {
		return nil, err
	}

	return f.Interface(), nil
}

func (p *proxy) set(t map[string]interface{}, k string, v, recv interface{}) (bool, error) {
	f, err := p.getField(t, k)
	if err != nil {
		return false, err
	}

	v = castNumberToGoType(f.Kind(), v)
	f.Set(reflect.ValueOf(v))
	return true, nil
}

func (p *proxy) getField(t map[string]interface{}, k string) (reflect.Value, error) {
	var r reflect.Value
	ptr, ok := t[goStructPtrProp].(unsafe.Pointer)
	if !ok {
		return r, UnexpectedPointer
	}

	v := reflect.ValueOf(p.structs[ptr]).Elem()
	switch v.Kind() {
	case reflect.Struct:
		k = strings.Title(k)
		r = v.FieldByName(k)
		if !r.IsValid() {
			r = reflect.ValueOf(p.structs[ptr]).MethodByName(k)
		}
	case reflect.Map:
		vk := reflect.ValueOf(k)
		r = v.MapIndex(vk)
	}

	if !r.IsValid() {
		return r, UndefinedProperty
	}

	return r, nil
}
