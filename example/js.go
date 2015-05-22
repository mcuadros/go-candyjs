package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mcuadros/tba"
)

var ctx *tba.Context

func main() {
	ctx = tba.NewContext()

	ctx.RegisterFunc(multiply)
	ctx.RegisterFunc(sprintf)
	ctx.RegisterFunc(include)
	ctx.RegisterFunc(require)
	ctx.RegisterFunc(swap)
	ctx.RegisterInstance("console", &Log{})
	ctx.EvalFile("js.js")

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		result := ctx.PevalString(input)
		fmt.Println(result)
	}
}

func require(filename string) error {
	code, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	ctx.PevalString(fmt.Sprintf(`(function() {
        %s
    })()`, code))

	return nil
}

func swap(x, y int) (int, int) {
	return y, x
}

func include(filename string) {
	ctx.EvalFile(filename)
}

func multiply(a, b int) int {
	return a * b
}

func sprintf(format string, str ...interface{}) string {
	return fmt.Sprintf(format, str...)
}

type Log struct{}

func (l *Log) Log(str ...interface{}) {
	fmt.Println(str...)
}
