package main

import (
	"fmt"
	"os"

	"github.com/mcuadros/go-candyjs"
)

//go:generate candyjs import time
//go:generate candyjs import net/http
//go:generate candyjs import io/ioutil
//go:generate candyjs import github.com/gin-gonic/gin
func main() {
	fmt.Printf("Executing %q\n", os.Args[1])

	ctx := candyjs.NewContext()
	err := ctx.PevalFile(os.Args[1])
	if err != nil {
		fmt.Println("Error", err)
	}
}
