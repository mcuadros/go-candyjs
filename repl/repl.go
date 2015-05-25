package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mcuadros/tba/engine/runtime"

	"github.com/agtorre/gocolorize"
	"github.com/bobappleyard/readline"
	"github.com/olebedev/go-duktape"
)

var (
	prompt   = "> "
	okPaint  = gocolorize.NewColor("green:black").Paint
	errPaint = gocolorize.NewColor("red:black").Paint
)

func main() {
	ctx := runtime.GetContext()

	status := okPaint
	if len(os.Args) > 1 {
		if err := ctx.PevalFile(os.Args[1]); err != nil {
			status = errPaint
			fmt.Println(err.(*duktape.Error).Stack)
		}
	}

	for {
		input, err := readline.String(status(prompt))
		if err != nil {
			if err != io.EOF {
				fmt.Println("error: ", err)
			}
			break
		}

		if err := ctx.PevalString(input); err != nil {
			status = errPaint
			fmt.Println(err.(*duktape.Error).Stack)
		} else {
			readline.AddHistory(input)

			status = okPaint
			result := ctx.JsonEncode(-1)
			if len(result) != 0 {
				fmt.Println(result)
			}
		}
	}
}
