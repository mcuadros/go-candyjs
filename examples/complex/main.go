package main

import (
	"fmt"
	"os"

	"github.com/mcuadros/go-candyjs"
)

//go:generate candyjs import net/http
//go:generate candyjs import io/ioutil
func main() {
	script := os.Args[1]
	fmt.Printf("Executing %q\n", script)

	ctx := candyjs.NewContext()
	ctx.PevalFile(script)
}
