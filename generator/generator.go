package generator

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"io"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"

	appenvmarkers "github.com/holyhope/appenv-generator/markers"
)

var (
	keyFieldMarker      = markers.Must(markers.MakeDefinition(appenvmarkers.EnvironmentVariableName, markers.DescribesField, ""))
	ignoreFieldMarker   = markers.Must(markers.MakeDefinition(appenvmarkers.EnvironmentIgnore, markers.DescribesField, true))
	embeddedFieldMarker = markers.Must(markers.MakeDefinition(appenvmarkers.EmbeddedEnvironmentVariable, markers.DescribesField, true))
)

// +controllertools:marker:generateHelp

// Generator generates code containing ShallowCopy method implementations.
type Generator struct{}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	if err := markers.RegisterAll(into, keyFieldMarker, ignoreFieldMarker, embeddedFieldMarker); err != nil {
		return errors.Wrap(err, "cannot register markers")
	}

	into.AddHelp(
		keyFieldMarker,
		markers.SimpleHelp("object", "The name of the environment variable to inject into the container"),
	)

	return nil
}

func (Generator) Generate(ctx *genall.GenerationContext) error {
	// loop through the loaded packages
	for _, root := range ctx.Roots {
		ctx.Checker.Check(root, func(node ast.Node) bool {
			// ignore interfaces
			_, isIface := node.(*ast.InterfaceType)
			return !isIface
		})

		code := jen.NewFile(root.Name)
		notNil := false

		if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
			if shouldTakeCare(root, info) {
				result, err := GenerateCode(root, info)
				if err != nil {
					root.AddError(errors.Wrap(err, "cannot generate code"))

					return
				}

				code.Add(result)

				notNil = true
			}
		}); err != nil {
			root.AddError(err)
			return nil
		}

		if notNil {
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

const (
	methodeName = "GetApplicationEnvironments"
)

func GenerateCode(pkg *loader.Package, info *markers.TypeInfo) (*jen.Statement, error) {
	var finalErr error

	result := jen.Func().
		Params(jen.Id("o").Op("*").Id(info.Name)).
		Id(methodeName).
		Params(jen.Id("ctx").Id("context.Context")).
		Params(jen.Index().Qual("k8s.io/api/core/v1", "EnvVar")).
		BlockFunc(func(g *jen.Group) {
			const variableName = "envs"

			staticStatements := []jen.Code{}
			statementstoAppend := []func(*jen.Statement){}

			for _, field := range info.Fields {
				ignore := field.Markers.Get(appenvmarkers.EnvironmentIgnore)

				if ignore != nil {
					continue
				}

				staticStatement, postStatement, err := GenerateCodeWithField(pkg, variableName, field)
				if err != nil {
					finalErr = err
				}

				if staticStatement != nil {
					staticStatements = append(staticStatements, staticStatement)
				}
				if postStatement != nil {
					statementstoAppend = append(statementstoAppend, postStatement)
				}
			}

			if len(statementstoAppend) == 0 {
				if len(staticStatements) == 0 {
					g.Return(jen.Nil())
					return
				}

				g.Return(jen.Index().Qual("k8s.io/api/core/v1", "EnvVar").Values(staticStatements...))
				return
			}

			g.Id(variableName).Op(":=").Index().Qual("k8s.io/api/core/v1", "EnvVar").Values(staticStatements...)

			for _, statement := range statementstoAppend {
				g.Do(statement)
			}

			g.Return(jen.Id(variableName))
		}).Line()

	return result, finalErr
}

func GenerateCodeWithField(pkg *loader.Package, variableName string, field markers.FieldInfo) (*jen.Statement, func(*jen.Statement), error) {
	ftype := pkg.TypesInfo.TypeOf(field.RawField.Type)

	if ftype == types.Typ[types.Invalid] {
		return nil, nil, nil
	}

	depth := 0
	for {
		ftyped, ok := ftype.(*types.Pointer)
		if ok {
			ftype = ftyped.Elem()
		} else {
			break
		}
	}

	switch ftyped := ftype.(type) {
	case *types.Basic:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			value := jen.Empty()

			switch ftyped.Kind() {
			case types.String:
				for i := 0; i < depth; i++ {
					value.Add(jen.Op("*"))
				}
				value.Add(jen.Id("o").Dot(field.Name))
			default:
				value.Add(jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Id("o").Dot(field.Name)))
			}

			return jen.Values(jen.Dict{
				jen.Id("Name"):  jen.Lit(envVarName),
				jen.Id("Value"): value,
			}), nil, nil
		}

		return nil, nil, nil

	case *types.Struct, *types.Interface, *types.Named:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			return jen.Values(jen.Dict{
				jen.Id("Name"):  jen.Lit(envVarName),
				jen.Id("Value"): jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Id("o").Dot(field.Name)),
			}), nil, nil
		}

		// TODO: jen.If to ensure the field implements the interface

		// TODO: How to handle embedded structure ?
		// Get the field name from an other way
		if field.Name != "" {
			//embedded := field.Markers.Get(appenvmarkers.EmbeddedEnvironmentVariable)
			//if embedded != nil {
			return nil, func(s *jen.Statement) {
				s.If(
					jen.List(jen.Id("a"), jen.Id("ok")).Op(":=").Id("o").Dot(field.Name).Assert(jen.Qual("github.com/holyhope/appenv-generator/v1", "ApplicationWithEnvironment")),
					jen.Id("ok"),
				).Block(
					jen.Id(variableName).Op("=").Append(
						jen.Id(variableName),
						jen.Id("a").Dot(methodeName).Call(jen.Id("ctx")).Op("..."),
					),
				)
			}, nil
			//}
		}

		return nil, nil, nil
	default:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			return nil, nil, errors.Errorf("type not yet supported: %v", ftyped)
		}

		embedded := field.Markers.Get(appenvmarkers.EmbeddedEnvironmentVariable)
		if embedded != nil {
			return nil, nil, errors.Errorf("type not yet supported: %v", ftyped)
		}

		return nil, nil, nil
		/*
			case *types.Struct, *ast.InterfaceType:
				envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)

				if envVarName != nil {
					s.Values(jen.Dict{
						jen.Id("Name"):  jen.Lit(envVarName),
						jen.Id("Value"): jen.Id("o").Dot(field.Name).Dot(methodeName).Call(jen.Id("ctx")),
					})
				}

				panic(fmt.Errorf("not supported for field struct"))

			case *ast.ArrayType:
				panic(fmt.Errorf("not supported for field array"))

			case *ast.MapType:
				panic(fmt.Errorf("not supported for field map"))
		*/
	}
}

// shouldBeCopied checks if we're supposed to make shallowcopy methods on the given type.
//
// This is the case if it's exported *and* either:
// - has a partial manual ShallowCopy implementation (in which case we fill in the rest)
// - aliases to a non-basic type eventually
// - is a struct
func shouldTakeCare(pkg *loader.Package, info *markers.TypeInfo) bool {
	if !ast.IsExported(info.Name) {
		return false
	}

	typeInfo := pkg.TypesInfo.TypeOf(info.RawSpec.Name)
	if typeInfo == types.Typ[types.Invalid] {
		pkg.AddError(loader.ErrFromNode(fmt.Errorf("unknown type %s", info.Name), info.RawSpec))
		return false
	}

	// according to gengo, everything named is an alias, except for an alias to a pointer,
	// which is just a pointer, afaict.  Just roll with it.
	if asPtr, isPtr := typeInfo.(*types.Named).Underlying().(*types.Pointer); isPtr {
		typeInfo = asPtr
	}

	lastType := typeInfo
	if _, isNamed := typeInfo.(*types.Named); isNamed {
		// if it has a manual shallowcopy, we're fine
		if HasGetApplicationEnvironmentsMethod(pkg, typeInfo) {
			return false
		}

		for underlyingType := typeInfo.Underlying(); underlyingType != lastType; lastType, underlyingType = underlyingType, underlyingType.Underlying() {
			// if it has a manual shallowcopy, we're fine
			if HasGetApplicationEnvironmentsMethod(pkg, underlyingType) {
				return true
			}

			// aliases to other things besides basics need copy methods
			// (basics can be straight-up shallow-copied)
			if _, isBasic := underlyingType.(*types.Basic); !isBasic {
				return true
			}
		}
	}

	// structs are the only thing that's not a basic that's copiable by default
	_, isStruct := lastType.(*types.Struct)
	return isStruct
}

func IsMarked(pkg *loader.Package, info *markers.TypeInfo) bool {
	return true
}

func HasGetApplicationEnvironmentsMethod(pkg *loader.Package, typeInfo types.Type) bool {
	method, ind, _ := types.LookupFieldOrMethod(typeInfo, true /* check pointers too */, pkg.Types, "GetApplicationEnvironments")
	if len(ind) != 1 {
		// ignore embedded methods
		return false
	}
	if method == nil {
		return false
	}

	methodSig := method.Type().(*types.Signature)
	if methodSig.Params() != nil && methodSig.Params().Len() != 1 {
		return false
	}
	if methodSig.Results() == nil || methodSig.Results().Len() != 2 {
		return false
	}

	return true
}

const (
	outputFile = "zz_generated.appenv.go"
)

// writeFormatted outputs the given code, after gofmt-ing it.  If we couldn't gofmt,
// we write the unformatted code for debugging purposes.
func writeOut(ctx *genall.GenerationContext, root *loader.Package, outBytes []byte) {
	outputFile, err := ctx.Open(root, outputFile)
	if err != nil {
		root.AddError(errors.Wrapf(err, "cannot open file %s", outputFile))

		return
	}

	defer outputFile.Close()

	n, err := outputFile.Write(outBytes)
	if err != nil {
		root.AddError(errors.Wrap(err, "cannot write output"))

		return
	}

	if n < len(outBytes) {
		root.AddError(io.ErrShortWrite)
	}
}
