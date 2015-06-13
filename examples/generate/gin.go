package main

import (
	"fmt"

	"github.com/mcuadros/go-candyjs"
)

//go:generate candyjs import time
//go:generate candyjs import github.com/gin-gonic/gin
func main() {
	ctx := candyjs.NewContext()
	ctx.PushPackage("time", "time")
	ctx.PushPackage("github.com/gin-gonic/gin", "gin")

	err := ctx.PevalString(`
        r = gin.Default()
        r.GET("/ping", CandyJS.proxy(function(c) {
            future = time.Date(2015, 10, 21, 4, 29 ,0, 0, time.UTC)
            now = time.Now()

            c.JSON(200, {
                "future": future.String(),
                "now": now.String(),
                "str": "Back to the Future day is on: " + future.sub(now) + " nsecs!",
            })
        }))
        r.Run(":8080") 
    `)

	fmt.Println(err)
}
