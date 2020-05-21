package generator

import (
	"bytes"
	"go/ast"
	"go/format"
	"sync"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

// Generator generates code containing ShallowCopy method implementations.
type Generator struct {
	types map[implementsKey]implementsValue
}

func (g Generator) Generate(ctx *genall.GenerationContext) error {
	// loop through the loaded packages
	for _, root := range ctx.Roots {
		g.types = map[implementsKey]implementsValue{}

		ctx.Checker.Check(root, func(node ast.Node) bool {
			// ignore interfaces
			_, isIface := node.(*ast.InterfaceType)
			return !isIface
		})

		statements := map[*markers.TypeInfo][]func(*jen.Statement){}
		lock := sync.Mutex{}

		if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
			if shouldTakeCare(root, info) {
				result, err := g.GenerateCode(root, info)
				if err != nil {
					root.AddError(errors.Wrap(err, "cannot generate code"))

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

			for typeInfo, statements := range statements {
				if g.doesTypeImplementsAppEnv(root, typeInfo) {
					for _, statement := range statements {
						code.Do(statement).Line()
					}
				}
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

			writeOut(ctx, root, outContents)
		}
	}

	return nil
}
