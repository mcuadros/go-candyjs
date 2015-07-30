package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mcuadros/go-candyjs"
)

func main() {
	ctx := candyjs.NewContext()
	t := ctx.NewTransaction()
	t.PushGoFunction("handleFunc", http.HandleFunc)
	t.PushGoFunction("listenAndServe", http.ListenAndServe)
	t.PushGoFunction("writeString", io.WriteString)

	/*
		ctx.PushGlobalGoFunction(candyjs.NoTransaction, "handleFunc", http.HandleFunc)
		ctx.PushGlobalGoFunction(candyjs.NoTransaction, "listenAndServe", http.ListenAndServe)
		ctx.PushGlobalGoFunction(candyjs.NoTransaction, "writeString", io.WriteString)
	*/
	err := ctx.EvalString(candyjs.NoTransaction, `
        handler = function(writer, request) {
            writeString(writer, "Hello from CandyJS!")
        }

        handleFunc("/", CandyJS.proxy(handler))
        listenAndServe(":8000", null)
  `)

	fmt.Println(err)
}
