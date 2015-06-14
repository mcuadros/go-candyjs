package candyjs

import "errors"

type PackagePusher func(ctx *Context)

var PackageNotFound = errors.New("Unable to find the requested package")
var pushers = make(map[string]PackagePusher, 0)

func RegisterPackagePusher(pckgName string, f PackagePusher) {
	pushers[pckgName] = f
}

func (ctx *Context) PushGlobalPackage(pckgName, alias string) error {
	ctx.PushGlobalObject()

	err := ctx.pushPackage(pckgName)
	if err != nil {
		return err
	}

	ctx.PutPropString(-2, alias)
	ctx.Pop()

	return nil
}

func (ctx *Context) pushPackage(pckgName string) error {
	f, ok := pushers[pckgName]
	if !ok {
		return PackageNotFound
	}

	f(ctx)

	return nil
}
