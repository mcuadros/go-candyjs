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
	ctx.ErrorHandler = func(err error) {
		fmt.Println(errPaint(err.Error()))
	}

	status := okPaint
	if len(os.Args) > 1 {
		if ctx.PevalFile(os.Args[1]) != 0 {
			fmt.Println("error evalutating", os.Args[1])
			status = errPaint
		}
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(status(prompt))
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if ctx.PevalString(input) == 0 {
			status = okPaint
		} else {
			status = errPaint
		}
	}
}
