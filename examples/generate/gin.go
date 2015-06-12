package main

import (
	"fmt"

	"github.com/mcuadros/go-candyjs"
)

//go:generate candyjs import github.com/gin-gonic/gin
func main() {
	ctx := candyjs.NewContext()
	ctx.PushPackage("github.com/gin-gonic/gin", "gin")

	err := ctx.PevalString(`
        r = gin.Default()
        r.GET("/ping", CandyJS.proxy(function(c) {
            c.String(200, "pong")
        }))
        r.Run(":8080") 
    `)

	fmt.Println(err)
}
