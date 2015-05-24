package duktape

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type DuktapeSuite struct {
	ctx *Context
}

var _ = Suite(&DuktapeSuite{})

func (s *DuktapeSuite) SetUpTest(c *C) {
	s.ctx = NewContext()
}

func (s *DuktapeSuite) TestRegisterFunc_String(c *C) {
	var called interface{}
	s.ctx.RegisterFunc("test_in_string", func(s string) {
		called = s
	})

	s.ctx.EvalString("test_in_string('foo')")
	c.Assert(called, Equals, "foo")
}

func (s *DuktapeSuite) TestRegisterFunc_Int(c *C) {
	var called interface{}
	s.ctx.RegisterFunc("test_in_int", func(i int) {
		called = i
	})

	s.ctx.EvalString("test_in_int(42)")
	c.Assert(called, Equals, 42)
}

func (s *DuktapeSuite) TestRegisterFunc_Float(c *C) {
	var called64 interface{}
	var called32 interface{}
	s.ctx.RegisterFunc("test_in_float", func(f64 float64, f32 float32) {
		called64 = f64
		called32 = f32
	})

	s.ctx.EvalString("test_in_float(42, 42)")
	c.Assert(called64, Equals, 42.0)
	c.Assert(called32, Equals, float32(42.0))
}

func (s *DuktapeSuite) TestRegisterFunc_Bool(c *C) {
	var called interface{}
	s.ctx.RegisterFunc("test_in_bool", func(b bool) {
		called = b
	})

	s.ctx.EvalString("test_in_bool(true)")
	c.Assert(called, Equals, true)
}

func (s *DuktapeSuite) TestRegisterFunc_Interface(c *C) {
	var called interface{}
	s.ctx.RegisterFunc("test_in_interface", func(i interface{}) {
		called = i
	})

	s.ctx.EvalString("test_in_interface('qux')")
	c.Assert(called, Equals, "qux")
}

func (s *DuktapeSuite) TestRegisterFunc_Slice(c *C) {
	var called interface{}
	s.ctx.RegisterFunc("test_in_slice", func(s []interface{}) {
		called = s
	})

	s.ctx.EvalString("test_in_slice(['foo', 42])")
	c.Assert(called, DeepEquals, []interface{}{"foo", 42})
}

func (s *DuktapeSuite) TestRegisterFunc_Map(c *C) {
	var called interface{}
	s.ctx.RegisterFunc("test_in_map", func(s map[string]interface{}) {
		called = s
	})

	s.ctx.EvalString("test_in_map({foo: 42, qux: 21})")
	c.Assert(called, DeepEquals, map[string]interface{}{"foo": 42, "qux": 21})
}

func (s *DuktapeSuite) TestRegisterFunc_Variadic(c *C) {
	var calledA interface{}
	var calledB interface{}
	s.ctx.RegisterFunc("test_in_variadic", func(s string, is ...int) {
		calledA = s
		calledB = is
	})

	s.ctx.EvalString("test_in_variadic('foo', 21, 42)")
	c.Assert(calledA, DeepEquals, "foo")
	c.Assert(calledB, DeepEquals, []int{21, 42})
}

func (s *DuktapeSuite) TearDownTest(c *C) {
	s.ctx.DestroyHeap()
}
