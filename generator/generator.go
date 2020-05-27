package generator

import (
	"bytes"
	"go/ast"
	"go/format"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"

	"github.com/holyhope/appenv-generator/generator/codegen/v2"
)

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

// Generator generates code containing ShallowCopy method implementations.
type Generator struct {
	// HeaderFile specifies the header text (e.g. license) to prepend to generated files.
	HeaderFile string `marker:",optional"`
	// Year specifies the year to substitute for " YEAR" in the header file.
	Year string `marker:",optional"`
}

func (g Generator) Generate(ctx *genall.GenerationContext) error {
	codeGen := codegen.NewCodeGen()

	// loop through the loaded packages
	for _, root := range ctx.Roots {
		ctx.Checker.Check(root, func(node ast.Node) bool {
			// ignore interfaces
			_, isIface := node.(*ast.InterfaceType)
			return !isIface
		})

		statements := map[*markers.TypeInfo][]func(*jen.Statement){}
		lock := sync.Mutex{}

		if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
			if shouldTakeCare(root, info) {
				result, err := codeGen.GenerateCode(root, info)
				if err != nil {
					root.AddError(errors.Wrapf(err, "cannot generate code for %s.%s", root.PkgPath, info.Name))

					return
				}

				lock.Lock()
				defer lock.Unlock()

				statements[info] = append(statements[info], result)
			}
		}); err != nil {
			root.AddError(err)
			return nil
		}

		if len(statements) > 0 {
			code := jen.NewFile(root.Name)

			removeFile := true

			for typeInfo, statements := range statements {
				if codeGen.TypeImplementsAppEnv(root, typeInfo) {
					for _, statement := range statements {
						code.Do(statement).Line()
						removeFile = false
					}
				}
			}

			if removeFile {
				if len(root.GoFiles) == 0 {
					root.AddError(errors.New("cannot determine file to remove"))

					return nil
				}
				base := path.Dir(root.GoFiles[0])
				fullPath := path.Join(base, outputFile)
				err := os.Remove(fullPath)
				if err != nil && !os.IsNotExist(err) {
					root.AddError(errors.Wrap(err, "cannot remove empty file"))
				}

				continue
			}

			var b bytes.Buffer

			err := code.Render(&b)
			if err != nil {
				root.AddError(errors.Wrap(err, "cannot render generated code"))

				return nil
			}

			outContents, err := format.Source(b.Bytes())
			if err != nil {
				root.AddError(errors.Wrap(err, "cannot format rendered code"))

				return nil
			}

			g.writeOut(ctx, root, outContents)
		}
	}

	return nil
}

func (g Generator) getHeader(ctx *genall.GenerationContext) (string, error) {
	if g.HeaderFile == "" {
		return "", nil
	}

	headerBytes, err := ctx.ReadFile(g.HeaderFile)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(string(headerBytes), " YEAR", " "+g.Year) + "\n\n", nil
}
