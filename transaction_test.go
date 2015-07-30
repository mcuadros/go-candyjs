package candyjs

import (
	"sync"
	"time"

	. "gopkg.in/check.v1"
)

func (s *CandySuite) TestTransactionPushStruct(c *C) {
	parallelize(c, func(*C) {
		t := s.ctx.NewTransaction()
		t.PushInterface("foo", &MyStruct{
			Int:     42,
			Float64: 21.0,
			Nested:  &MyStruct{Int: 21},
		})
	})
}

func parallelize(c *C, f func(*C)) {
	count := 1000
	var wg sync.WaitGroup
	wg.Add(count)

	sleep := time.Millisecond * 5
	for i := 0; i < count; i++ {
		start := time.Now()
		go func(s time.Duration) {
			time.Sleep(s)

			for j := 0; j < 100; j++ {
				f(c)
			}

			wg.Done()
		}(sleep)

		sleep -= time.Now().Sub(start)
	}

	wg.Wait()
}
