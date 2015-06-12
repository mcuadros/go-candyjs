package candyjs

import "errors"

type PackagePusher func(ctx *Context, alias string)

var PackageNotFound = errors.New("Unable to find the requested package")
var pushers = make(map[string]PackagePusher, 0)

func RegisterPackagePusher(pckgName string, f PackagePusher) {
	pushers[pckgName] = f
}

func (ctx *Context) PushPackage(pckgName, alias string) error {
	f, ok := pushers[pckgName]
	if !ok {
		return PackageNotFound
	}

	f(ctx, alias)

	return nil
}
