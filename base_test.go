package candyjs

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type DuktapeSuite struct {
	ctx    *Context
	stored interface{}
}

var _ = Suite(&DuktapeSuite{})

func (s *DuktapeSuite) SetUpTest(c *C) {
	s.ctx = NewContext()
	s.stored = nil
	s.ctx.PushGlobalGoFunction("store", func(value interface{}) {
		s.stored = value
	})
}

func (s *DuktapeSuite) TestSetRequireFunction(c *C) {
	s.ctx.SetRequireFunction(func(id string, a ...interface{}) string {
		return fmt.Sprintf(`exports.store = function () { store("%s"); };`, id)
	})

	c.Assert(s.ctx.PevalString("require('foo').store()"), IsNil)
	c.Assert(s.stored, Equals, "foo")
}

func (s *DuktapeSuite) TestProxyGlobalInterface_Integration(c *C) {
	now := time.Now()
	after := now.Add(time.Millisecond)

	s.ctx.ProxyGlobalInterface("a", now)
	s.ctx.ProxyGlobalInterface("b", after)

	s.ctx.PevalString(`store(b.sub(a))`)
	c.Assert(s.stored, Equals, 1000000.0)
}

func (s *DuktapeSuite) TestProxyGlobalInterface_GetMap(c *C) {
	s.ctx.ProxyGlobalInterface("test", &map[string]int{"foo": 42})

	s.ctx.PevalString(`store(test.foo)`)
	c.Assert(s.stored, Equals, 42.0)
}

func (s *DuktapeSuite) TestProxyGlobalInterface_GetPtr(c *C) {
	s.ctx.ProxyGlobalInterface("test", &MyStruct{Int: 42})

	s.ctx.PevalString(`store(test.int)`)
	c.Assert(s.stored, Equals, 42.0)

	s.ctx.PevalString(`try { x = test.baz; } catch(err) { store(true); }`)
	c.Assert(s.stored, Equals, true)
}

func (s *DuktapeSuite) TestProxyGlobalInterface_Set(c *C) {
	s.ctx.ProxyGlobalInterface("test", &MyStruct{Int: 42})

	s.ctx.PevalString(`test.int = 21; store(test.int)`)
	c.Assert(s.stored, Equals, 21.0)

	s.ctx.PevalString(`try { test.baz = 21; } catch(err) { store(true); }`)
	c.Assert(s.stored, Equals, true)
}

func (s *DuktapeSuite) TestProxyGlobalInterface(c *C) {
	s.ctx.PushGlobalObject()
	s.ctx.PushObject()
	s.ctx.ProxyInterface(&MyStruct{Int: 142})
	s.ctx.PutPropString(-2, "obj")
	s.ctx.PutPropString(-2, "foo")
	s.ctx.Pop()

	err := s.ctx.PevalString(`store(foo.obj.int)`)
	c.Assert(err, IsNil)
	c.Assert(s.stored, Equals, 142.0)
}

func (s *DuktapeSuite) TestProxyGlobalInterface_Has(c *C) {
	s.ctx.ProxyGlobalInterface("test", &MyStruct{})
	s.ctx.PevalString(`store("int" in test)`)
	c.Assert(s.stored, Equals, true)

	s.ctx.PevalString(`store("qux" in test)`)
	c.Assert(s.stored, Equals, false)
}

func (s *DuktapeSuite) TestProxyGlobalInterface_Nested(c *C) {
	s.ctx.ProxyGlobalInterface("test", &MyStruct{
		Int:     42,
		Float64: 21.0,
		Nested:  &MyStruct{Int: 21},
	})

	c.Assert(s.ctx.PevalString(`store([
		test.int,
      	test.string(),
	    test.multiply(2),
	    test.nested.int,
	    test.nested.multiply(3)
	])`), IsNil)

	c.Assert(s.stored, DeepEquals, []interface{}{42.0, "qux", 84.0, 21.0, 63.0})
}

func (s *DuktapeSuite) TestPushGlobalValueInt(c *C) {
	s.ctx.PushGlobalValue("test", reflect.ValueOf(42))
	c.Assert(s.ctx.PevalString(`store(test)`), IsNil)
	c.Assert(s.stored, Equals, 42.0)
}

func (s *DuktapeSuite) TestPushGlobalValueFloat(c *C) {
	s.ctx.PushGlobalValue("test", reflect.ValueOf(42.2))
	c.Assert(s.ctx.PevalString(`store(test)`), IsNil)
	c.Assert(s.stored, Equals, 42.2)
}

func (s *DuktapeSuite) TestPushGlobalValueString(c *C) {
	s.ctx.PushGlobalValue("test", reflect.ValueOf("foo"))
	c.Assert(s.ctx.PevalString(`store(test)`), IsNil)
	c.Assert(s.stored, Equals, "foo")
}

func (s *DuktapeSuite) TestPushGlobalValueStruct(c *C) {
	s.ctx.PushGlobalValue("test", reflect.ValueOf(MyStruct{Int: 42}))
	c.Assert(s.ctx.PevalString(`store(test.int)`), IsNil)
	c.Assert(s.stored, Equals, 42.0)
}

func (s *DuktapeSuite) TestPushGlobalValueStructPtr(c *C) {
	s.ctx.PushGlobalValue("test", reflect.ValueOf(&MyStruct{Int: 42}))
	c.Assert(s.ctx.PevalString(`store(test.int)`), IsNil)
	c.Assert(s.stored, Equals, 42.0)
}

func (s *DuktapeSuite) TestPushGlobalValues(c *C) {
	s.ctx.PushGlobalValues("test", []reflect.Value{
		reflect.ValueOf("foo"), reflect.ValueOf("qux"),
	})

	c.Assert(s.ctx.PevalString(`store(test)`), IsNil)
	c.Assert(s.stored, DeepEquals, []interface{}{"foo", "qux"})
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_String(c *C) {
	var called interface{}
	s.ctx.PushGlobalGoFunction("test_in_string", func(s string) {
		called = s
	})

	s.ctx.EvalString("test_in_string('foo')")
	c.Assert(called, Equals, "foo")
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Int(c *C) {
	var ri, ri8, ri16, ri32, ri64 interface{}
	s.ctx.PushGlobalGoFunction("test_in_int", func(i int, i8 int8, i16 int16, i32 int32, i64 int64) {
		ri = i
		ri8 = i8
		ri16 = i16
		ri32 = i32
		ri64 = i64
	})

	s.ctx.EvalString("test_in_int(42, 8, 16, 32, 64)")
	c.Assert(ri, Equals, 42)
	c.Assert(ri8, Equals, int8(8))
	c.Assert(ri16, Equals, int16(16))
	c.Assert(ri32, Equals, int32(32))
	c.Assert(ri64, Equals, int64(64))
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Uint(c *C) {
	var ri, ri8, ri16, ri32, ri64 interface{}
	s.ctx.PushGlobalGoFunction("test_in_uint", func(i uint, i8 uint8, i16 uint16, i32 uint32, i64 uint64) {
		ri = i
		ri8 = i8
		ri16 = i16
		ri32 = i32
		ri64 = i64
	})

	s.ctx.EvalString("test_in_uint(42, 8, 16, 32, 64)")
	c.Assert(ri, Equals, uint(42))
	c.Assert(ri8, Equals, uint8(8))
	c.Assert(ri16, Equals, uint16(16))
	c.Assert(ri32, Equals, uint32(32))
	c.Assert(ri64, Equals, uint64(64))
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Float(c *C) {
	var called64 interface{}
	var called32 interface{}
	s.ctx.PushGlobalGoFunction("test_in_float", func(f64 float64, f32 float32) {
		called64 = f64
		called32 = f32
	})

	s.ctx.EvalString("test_in_float(42, 42)")
	c.Assert(called64, Equals, 42.0)
	c.Assert(called32, Equals, float32(42.0))
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Bool(c *C) {
	var called interface{}
	s.ctx.PushGlobalGoFunction("test_in_bool", func(b bool) {
		called = b
	})

	s.ctx.EvalString("test_in_bool(true)")
	c.Assert(called, Equals, true)
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Interface(c *C) {
	var called interface{}
	s.ctx.PushGlobalGoFunction("test", func(i interface{}) {
		called = i
	})

	s.ctx.EvalString("test('qux')")
	c.Assert(called, Equals, "qux")
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Struct(c *C) {
	var called *MyStruct
	s.ctx.PushGlobalGoFunction("test", func(m *MyStruct) {
		called = m
	})

	s.ctx.EvalString("test({'int':42})")
	c.Assert(called.Int, Equals, 42)
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Slice(c *C) {
	var called interface{}
	s.ctx.PushGlobalGoFunction("test_in_slice", func(s []interface{}) {
		called = s
	})

	s.ctx.EvalString("test_in_slice(['foo', 42])")
	c.Assert(called, DeepEquals, []interface{}{"foo", 42.0})
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Map(c *C) {
	var called interface{}
	s.ctx.PushGlobalGoFunction("test_in_map", func(s map[string]interface{}) {
		called = s
	})

	s.ctx.EvalString("test_in_map({foo: 42, qux: {bar: 'bar'}})")

	c.Assert(called, DeepEquals, map[string]interface{}{
		"foo": 42.0,
		"qux": map[string]interface{}{"bar": "bar"},
	})
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Nil(c *C) {
	var cm, cs, ci, cst interface{}
	s.ctx.PushGlobalGoFunction("test_nil", func(m map[string]interface{}, s []interface{}, i int, st string) {
		cm = m
		cs = s
		ci = i
		cst = st
	})

	s.ctx.EvalString("test_nil(null, null, null, null)")
	c.Assert(cm, DeepEquals, map[string]interface{}(nil))
	c.Assert(cs, DeepEquals, []interface{}(nil))
	c.Assert(ci, DeepEquals, 0)
	c.Assert(cst, DeepEquals, "")
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Optional(c *C) {
	var cm, cs, ci, cst interface{}
	s.ctx.PushGlobalGoFunction("test_optional", func(m map[string]interface{}, s []interface{}, i int, st string) {
		cm = m
		cs = s
		ci = i
		cst = st
	})

	s.ctx.EvalString("test_optional()")
	c.Assert(cm, DeepEquals, map[string]interface{}(nil))
	c.Assert(cs, DeepEquals, []interface{}(nil))
	c.Assert(ci, DeepEquals, 0)
	c.Assert(cst, DeepEquals, "")
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Variadic(c *C) {
	var calledA interface{}
	var calledB interface{}
	s.ctx.PushGlobalGoFunction("test_in_variadic", func(s string, is ...int) {
		calledA = s
		calledB = is
	})

	s.ctx.EvalString("test_in_variadic('foo', 21, 42)")
	c.Assert(calledA, DeepEquals, "foo")
	c.Assert(calledB, DeepEquals, []int{21, 42})
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_ReturnStruct(c *C) {
	s.ctx.PushGlobalGoFunction("test", func() *MyStruct {
		return &MyStruct{Int: 42}
	})

	c.Assert(s.ctx.PevalString("store(test().multiply(3))"), IsNil)
	c.Assert(s.stored, Equals, 126.0)
}

func (s *DuktapeSuite) TestPushGlobalGoFunction_Error(c *C) {
	s.ctx.PushGlobalGoFunction("test", func() (string, error) {
		return "foo", fmt.Errorf("foo")
	})

	c.Assert(s.ctx.PevalString(`
		try {
			test();
		} catch(err) {
			store(true);
		}
	`), IsNil)
	c.Assert(s.stored, Equals, true)
}

func (s *DuktapeSuite) TearDownTest(c *C) {
	s.ctx.DestroyHeap()
}

type MyStruct struct {
	Int     int
	Float64 float64
	Empty   *MyStruct
	Nested  *MyStruct
	Foo     []int
}

func (m *MyStruct) String() string {
	return "qux"
}

func (m *MyStruct) Multiply(x int) int {
	return m.Int * x
}
