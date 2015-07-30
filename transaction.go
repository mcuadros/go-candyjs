package candyjs

import "sync"

type transactionID int

var (
	transactionIDSeed transactionID
	noTransaction     transactionID = -1
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

func (t *Transaction) PushGlobalGoFunction(name string, f interface{}) (int, error) {
	return t.ctx.PushGlobalGoFunction(t.id, name, t.ctx.wrapFunction(f))
}

func (t *Transaction) PushGlobalType(name string, s interface{}) int {
	return t.ctx.PushGlobalType(t.id, name, s)
}

type mutex struct {
	lock    sync.Mutex
	current transactionID
}

func (m *mutex) Lock(id transactionID) {
	if id == -1 {
		return
	}

	if m.current == id {
		return
	}

	m.current = id
	m.lock.Lock()
}

func (m *mutex) Unlock(id transactionID) {
	if id == -1 {
		return
	}

	if m.current != id {
		panic("unlock of invalid id")
	}

	m.lock.Unlock()
	m.current = -1
}
