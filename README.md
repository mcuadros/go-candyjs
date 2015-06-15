go-candyjs [![Build Status](https://travis-ci.org/mcuadros/go-candyjs.png?branch=master)](https://travis-ci.org/mcuadros/go-candyjs) [![Coverage Status](https://coveralls.io/repos/mcuadros/go-candyjs/badge.svg?branch=master)](https://coveralls.io/r/mcuadros/go-candyjs?branch=master) [![GoDoc](http://godoc.org/github.com/mcuadros/go-candyjs?status.png)](http://godoc.org/github.com/mcuadros/go-candyjs) [![GitHub release](https://img.shields.io/github/release/mcuadros/go-candyjs.svg)](https://github.com/mcuadros/go-candyjs/releases)
==========

*CandyJS* is an intent of create a fully **transparent bridge between Go and the
JavaScript** engine [duktape](http://duktape.org/). Basicly is a syntax-sugar
library built it on top of [go-duktape](https://github.com/olebedev/go-duktape)
using reflection techniques.

#### ok but what for ...

build extensible applications that allow to the user **execute** arbitrary
**code** (let's say plugins) **without** the requirement of **compile** it. 

Demo
----

![asciicast](https://raw.githubusercontent.com/mcuadros/go-candyjs/master/examples/demo/cast.gif)

Features
--------
- Embeddable Ecmascript E5/E5.1 compliant engine ([duktape](http://duktape.org/)).
- Call Go function from JavaScript and vice versa.
- Transparent interface between Go structs and JavaScript.
- Import of Go packages into the JavaScript context.

Installation
------------

The recommended way to install go-candyjs is:

```
go get -u github.com/mcuadros/go-candyjs/...
```

> *CandyJS* includes a binary tool used by [go generate](http://blog.golang.org/generate),
please be sure that `$GOPATH/bin` is on your `$PATH`


Examples
--------

In this example a [`gin`](https://github.com/gin-gonic/gin) server is executed
and a small JSON is server. In CandyJS you can import Go packages directly if
they are [defined](https://github.com/mcuadros/go-candyjs/blob/master/examples/complex/main.go#L10:L13)
previously on the Go code. 

```js
var time = CandyJS.require('time');
var gin = CandyJS.require('github.com/gin-gonic/gin');

var engine = gin.default();
engine.get("/back", CandyJS.proxy(function(ctx) {
  var future = time.date(2015, 10, 21, 4, 29 ,0, 0, time.UTC);
  var now = time.now();

  ctx.json(200, {
    future: future.string(),
    now: now.string(),
    nsecs: future.sub(now)
  });
}));

engine.run(':8080');
```

The previous JS can be executed using this small piece of code:
```go
...
//go:generate candyjs import time
//go:generate candyjs import github.com/gin-gonic/gin
func main() {
	ctx := candyjs.NewContext()
	ctx.PevalFile("example.js")
}

```

Caveats
-------
Due to an [incompatibility](https://github.com/svaarala/duktape/issues/154#issuecomment-87077208) with Duktape's error handling system and Go, you can't throw errors from Go. All errors generated from Go functions are generic ones `error error (rc -100)`

License
-------

MIT, see [LICENSE](LICENSE)
