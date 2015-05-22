package runtime

import (
	"errors"
	"fmt"
	"time"
)

var NotImplemented = errors.New("not implemented")

const (
	ErrorLevel = "error"
	InfoLevel  = "info"
	TraceLevel = "trace"
	WarnLevel  = "warn"
)

type Console struct {
	count  int
	timers map[string]time.Time
}

func NewConsole() *Console {
	return &Console{
		timers: make(map[string]time.Time, 0),
	}
}

func (c *Console) Assert(...interface{}) error {
	return NotImplemented
}

func (c *Console) Count() int {
	c.count++
	return c.count
}

func (c *Console) Dir(...interface{}) error {
	return NotImplemented
}

func (c *Console) Error(str ...interface{}) {
	c.Log(append([]interface{}{ErrorLevel}, str...)...)
}

func (c *Console) Info(str ...interface{}) {
	c.Log(append([]interface{}{InfoLevel}, str...)...)
}

func (c *Console) Log(str ...interface{}) {
	fmt.Println(str...)
}

func (c *Console) Time(name string) {
	c.timers[name] = time.Now()
}

func (c *Console) TimeEnd(name string) {
	if _, ok := c.timers[name]; !ok {
		return
	}

	e := time.Now().Sub(c.timers[name])
	delete(c.timers, name)

	c.Log(fmt.Sprintf("%s: %s", name, e))
}

func (c *Console) Trace(str ...interface{}) {
	c.Log(append([]interface{}{TraceLevel}, str...)...)
}

func (c *Console) Warn(str ...interface{}) {
	c.Log(append([]interface{}{WarnLevel}, str...)...)
}
