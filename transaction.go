package candyjs

import "sync"

type transactionID int

var (
	transactionIDSeed transactionID
	noTransaction     transactionID = -1
	NoTransaction     transactionID = -1
)

type Transaction struct {
	id  transactionID
	ctx *Context
}

func newTransaction(ctx *Context) *Transaction {
	transactionIDSeed++
	return &Transaction{
		id:  transactionIDSeed,
		ctx: ctx,
	}
}

func (t *Transaction) PushStruct(name string, s interface{}) (int, error) {
	return t.ctx.PushGlobalStruct(t.id, name, s)
}

func (t *Transaction) PushInterface(name string, v interface{}) error {
	return t.ctx.PushGlobalInterface(t.id, name, v)
}

func (t *Transaction) PushGoFunction(name string, f interface{}) (int, error) {
	return t.ctx.PushGlobalGoFunction(t.id, name, f)
}

func (t *Transaction) PushType(name string, s interface{}) int {
	return t.ctx.PushGlobalType(t.id, name, s)
}

func (t *Transaction) EvalString(src string) error {
	return t.ctx.EvalString(t.id, src)
}

type mutex struct {
	lock    sync.Mutex
	locks   int
	current transactionID
}

func (m *mutex) Lock(id transactionID) {
	if id == -1 {
		return
	}

	if m.current == id {
		m.locks++
		return
	}

	m.lock.Lock()
	m.current = id
	m.locks++
}

func (m *mutex) Unlock(id transactionID) {
	if id == -1 {
		return
	}

	if m.current != id {
		panic("unlock of invalid id")
	}

	m.locks--
	if m.locks == 0 {
		m.lock.Unlock()
		m.current = -1
	}
}
