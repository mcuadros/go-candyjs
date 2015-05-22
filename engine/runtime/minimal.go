package runtime

import (
	"fmt"
	"io/ioutil"
)

func require(filename string) error {
	code, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	ctx.PevalString(fmt.Sprintf(`(function() {
        var exports = module.exports = {};
        %s
        return module.exports
    })()`, code))

	return nil
}

func include(filename string) {
	ctx.EvalFile(filename)
}
