package candyjs

import (
	. "gopkg.in/check.v1"
)

func (s *DuktapeSuite) TestProxy_Has(c *C) {
	c.Assert(p.has(&MyStruct{Int: 42}, "int"), Equals, true)
}

func (s *DuktapeSuite) TestProxy_Get(c *C) {
	v, err := p.get(&MyStruct{Int: 42}, "int", nil)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, 42)
}

func (s *DuktapeSuite) TestProxy_Set(c *C) {
	t := &MyStruct{Int: 21}

	setted, err := p.set(t, "int", 42.0, nil)
	c.Assert(err, IsNil)
	c.Assert(setted, Equals, true)

	v, err := p.get(t, "int", nil)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, 42)
}

func (s *DuktapeSuite) TestProxy_Enumerate(c *C) {
	keys, err := p.enumerate(&MyStruct{Int: 42})
	c.Assert(err, IsNil)
	c.Assert(
		keys,
		DeepEquals,
		[]string{"int", "float64", "empty", "nested", "foo", "multiply", "string"},
	)
}

func (s *DuktapeSuite) TestProxy_SetOnFunction(c *C) {
	setted, err := p.set(&MyStruct{Int: 21}, "string", 42.0, nil)
	c.Assert(err, IsNil)
	c.Assert(setted, Equals, false)
}

func (s *DuktapeSuite) TestProxy_Properties(c *C) {
	provider := [][]interface{}{
		{&MyStruct{Int: 32}, "int", 32},
		{MyStruct{Int: 42}, "int", 42},
		{map[string]int{"foo": 21}, "foo", 21},
		{&map[string]int{"foo": 42}, "foo", 42},
	}

	for _, v := range provider {
		s.testProxyProperties(c, v[0], v[1], v[2])
	}
}

func (s *DuktapeSuite) testProxyProperties(c *C, value, key, expected interface{}) {
	val, err := p.get(value, key.(string), nil)
	c.Assert(err, IsNil)
	c.Assert(val, Equals, expected)
}

func (s *DuktapeSuite) TestProxy_Functions(c *C) {
	provider := [][]interface{}{
		{&MyStruct{}, "string"},
		{&customMap{}, "functionWithPtr"},
		{customMap{}, "functionWithoutPtr"},
		{customInt(1), "functionWithoutPtr"},
	}

	for _, v := range provider {
		s.testProxyFunction(c, v[0], v[1])
	}
}

func (s *DuktapeSuite) testProxyFunction(c *C, value, key interface{}) {
	val, err := p.get(value, key.(string), nil)
	c.Assert(err, IsNil)
	c.Assert(val, NotNil)
}

type customInt int

func (c customInt) FunctionWithoutPtr() {}

type customMap map[string]int

func (c customMap) FunctionWithoutPtr() {}
func (c *customMap) FunctionWithPtr()   {}
