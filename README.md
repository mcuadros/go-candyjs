go-candyjs [![Build Status](https://travis-ci.org/mcuadros/go-candyjs.png?branch=master)](https://travis-ci.org/mcuadros/go-candyjs) [![Coverage Status](https://coveralls.io/repos/mcuadros/go-candyjs/badge.svg?branch=master)](https://coveralls.io/r/mcuadros/go-candyjs?branch=master) [![GoDoc](http://godoc.org/github.com/mcuadros/go-candyjs?status.png)](http://godoc.org/github.com/mcuadros/go-candyjs) [![GitHub release](https://img.shields.io/github/release/mcuadros/go-candyjs.svg)](https://github.com/mcuadros/go-candyjs/releases)
==========

*CandyJS* is an intent of create a fully transparent bridge between Go and the JavaScript engine [duktape](http://duktape.org/). Basicly is a syntax-sugar library built it on top of [go-duktape](https://github.com/olebedev/go-duktape) using reflection techniques.

Installation
------------

The recommended way to install go-candyjs

```
go get github.com/mcuadros/go-candyjs
```

Examples
--------

```go
package main

import (
    "time"

    "github.com/mcuadros/go-candyjs"
)

func main() {
    ctx := candyjs.NewContext()
    ctx.PushGlobalGoFunction("date", time.Date)
    ctx.PushGlobalGoFunction("now", time.Now)
    ctx.PushGlobalProxy("UTC", time.UTC)

    ctx.EvalString(`
        future = date(2015, 10, 21, 4, 29 ,0, 0, UTC)

        print("Back to the Future day is on: " + future.sub(now()) + " nsecs!")
    `)
}
```

Caveats
-------
- Due to an [incompatibility](https://github.com/svaarala/duktape/issues/154#issuecomment-87077208) with Duktape's error handling system and Go, you can't throw errors from Go. All errors generated from Go functions are generic ones `error error (rc -100)`

License
-------

MIT, see [LICENSE](LICENSE)
