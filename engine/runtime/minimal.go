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

	return ctx.PevalString(fmt.Sprintf(`(function() {
        var exports = module.exports = {};
        %s
        return module.exports
    })()`, code))
}

func include(filename string) error {
	return ctx.PevalFile(filename)
}
