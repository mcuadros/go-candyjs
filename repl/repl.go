package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mcuadros/tba/engine/runtime"

	"github.com/agtorre/gocolorize"
)

var (
	prompt   = "> "
	okPaint  = gocolorize.NewColor("green:black").Paint
	errPaint = gocolorize.NewColor("red:black").Paint
)

func main() {
	ctx := runtime.GetContext()
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(okPaint(prompt))
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if ctx.PevalString(input) == 0 {
			fmt.Printf(okPaint(prompt))
		} else {
			fmt.Printf(errPaint(prompt))
		}
	}
}
