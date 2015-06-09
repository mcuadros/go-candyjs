package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type CmdImport struct {
	Output string `short:"" long:"output" description:"output file name"`
	Alias  string `short:"" long:"alias" description:"packages alias"`

	Args struct {
		Package string `positional-arg-name:"package" description:"package to import"`
	} `positional-args:"yes"`
}

func (c *CmdImport) Execute(args []string) error {
	fmt.Printf("Processing %q as %q\n", c.Args.Package, c.Alias)

	objects, err := c.getObjects(c.Args.Package)
	if err != nil {
		fmt.Println(err)

		return err
	}

	fmt.Println(objects)

	return nil
}

func (c *CmdImport) getObjects(pkgName string) (map[string]*ast.Object, error) {
	pkgs, err := c.parserPackage(pkgName)
	if err != nil {
		return nil, err
	}

	var objects map[string]*ast.Object
	for _, pkg := range pkgs {
		pkgObjs := c.getPackageObjects(pkg)
		if len(pkgObjs) == 0 {
			continue
		} else if len(objects) != 0 && len(pkgObjs) != 0 {
			return nil, errors.New("Unable to process a dir with multiple packages")
		}

		objects = pkgObjs
	}

	return objects, nil
}

func (c *CmdImport) parserPackage(pkgName string) (map[string]*ast.Package, error) {
	dir, err := c.getPackagePath(pkgName)
	if err != nil {
		return nil, err
	}

	return parser.ParseDir(token.NewFileSet(), dir, nil, 0)
}

func (c *CmdImport) getPackageObjects(pkg *ast.Package) map[string]*ast.Object {
	objects := make(map[string]*ast.Object)

	for filename, f := range pkg.Files {
		if strings.HasSuffix(filepath.Base(filename), "_test.go") {
			continue
		}

		for name, object := range f.Scope.Objects {
			if ast.IsExported(name) {
				objects[name] = object
			}
		}

		fmt.Printf("Processed package file %q\n", filename)
	}

	return objects
}

func (c *CmdImport) getPackagePath(pkgName string) (string, error) {
	if pkgName == "" {
		return "", errors.New(fmt.Sprintf("invalid package name %q", pkgName))
	}

	for _, base := range []string{os.Getenv("GOPATH"), runtime.GOROOT()} {
		dir := filepath.Join(base, "src", pkgName)
		_, err := os.Stat(dir)
		if err == nil {
			return dir, nil
		}
	}

	return "", errors.New(fmt.Sprintf("package %q not found", pkgName))
}
