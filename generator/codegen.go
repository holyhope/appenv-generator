package generator

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

import (
	"go/types"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"

	appenvmarkers "github.com/holyhope/appenv-generator/markers"
)

func (g Generator) GenerateCode(pkg *loader.Package, info *markers.TypeInfo) (func(*jen.Statement), error) {
	const variableName = "envs"

	statement := jen.Empty()

	staticStatements := []jen.Code{}
	statementstoAppend := []func(*jen.Statement, *jen.Statement) func(*jen.Statement){}
	takeTimeCount := 0

	for _, field := range info.Fields {
		ignore := field.Markers.Get(appenvmarkers.EnvironmentIgnore)

		if ignore != nil {
			continue
		}

		staticStatement, postStatement, takeTime, err := g.GenerateCodeWithField(pkg, jen.Id(variableName), field)
		if err != nil {
			return nil, errors.Wrapf(err, "field %s", field.Name)
		}

		if takeTime {
			takeTimeCount++
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
			g.setTypeImplementsAppEnv(pkg, info, doesNotImplements)
			return nil, nil
		}
	}

	if takeTimeCount == 0 {
		g.setTypeImplementsAppEnv(pkg, info, doesImplements)
	} else {
		g.setTypeImplementsAppEnv(pkg, info, inheritImplementation)
	}

	if len(staticStatements) != 0 {
		g.setTypeImplementsAppEnv(pkg, info, doesImplements)
	}

	statement = jen.Id(variableName).Op(":=").Index().Qual("k8s.io/api/core/v1", "EnvVar").Values(staticStatements...)

	return func(s *jen.Statement) {
		s.Func().
			Params(jen.Id("o").Op("*").Id(info.Name)).
			Id(methodeName).
			Params(jen.Id("ctx").Qual("context", "Context")).
			Params(jen.Index().Qual("k8s.io/api/core/v1", "EnvVar"), jen.Id("error")).
			BlockFunc(func(g *jen.Group) {
				g.If(jen.Id("o").Op("==").Nil()).Block(jen.Return(jen.Nil(), jen.Nil()))
				g.Line()

				g.Add(statement)
				g.Line()

				var errGroup, lock *jen.Statement

				if takeTimeCount > 1 {
					// g, ctx := errgroup.WithContext(ctx)
					g.List(jen.Id("errGroup"), jen.Id("ctx")).Op(":=").Qual("golang.org/x/sync/errgroup", "WithContext").Call(jen.Id("ctx"))
					g.Var().Id("lock").Qual("sync", "Mutex")
					g.Line()

					errGroup = jen.Id("errGroup")
					lock = jen.Id("lock").Dot("Lock").Call().Line().Id("defer").Id("lock").Dot("Unlock").Call()
				}

				for _, statement := range statementstoAppend {
					g.Do(statement(errGroup, lock))
					g.Line()
				}

				errStatement := jen.Nil()

				if takeTimeCount > 1 {
					errStatement = jen.Id("errGroup").Dot("Wait").Call()
				}

				g.Return(jen.Id(variableName), errStatement)
			})
	}, nil
}

func (g Generator) GenerateCodeWithField(pkg *loader.Package, variable *jen.Statement, field markers.FieldInfo) (*jen.Statement, func(*jen.Statement, *jen.Statement) func(*jen.Statement), bool, error) {
	ftype := pkg.TypesInfo.TypeOf(field.RawField.Type)

	if ftype == types.Typ[types.Invalid] {
		return nil, nil, false, nil
	}

	isPointer := false
	if ftyped, ok := ftype.(*types.Pointer); ok {
		ftype = ftyped.Elem()
		isPointer = true
	}

	switch ftyped := ftype.(type) {
	case *types.Basic:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			value := jen.Empty()

			if isPointer {
				// Ensure field is not nil before dereferencing
				return nil, func(errGroup *jen.Statement, lock *jen.Statement) func(s *jen.Statement) {
					return func(s *jen.Statement) {
						s.
							If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
							Block(variable.Clone().Op("=").Append(
								variable,
								jen.Qual("k8s.io/api/core/v1", "EnvVar").Values(jen.Dict{
									jen.Id("Name"):  jen.Lit(envVarName),
									jen.Id("Value"): value.Add(jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Op("*").Id("o").Dot(field.Name))),
								}),
							))
					}
				}, false, nil
			}

			switch ftyped.Kind() {
			case types.String:
				value.Add(jen.Id("o").Dot(field.Name))
			default:
				value.Add(jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Id("o").Dot(field.Name)))
			}

			return jen.Values(jen.Dict{
				jen.Id("Name"):  jen.Lit(envVarName),
				jen.Id("Value"): value,
			}), nil, false, nil
		}

		return nil, nil, false, nil

	case *types.Struct, *types.Interface, *types.Named:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			return jen.Values(jen.Dict{
				jen.Id("Name"):  jen.Lit(envVarName),
				jen.Id("Value"): jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Id("o").Dot(field.Name)),
			}), nil, false, nil
		}

		fieldName := field.Name

		if fieldName == "" {
			fullNameParts := strings.Split(ftype.String(), ".")
			fieldName = fullNameParts[len(fullNameParts)-1]
		}

		//embedded := field.Markers.Get(appenvmarkers.EmbeddedEnvironmentVariable)
		//if embedded != nil {
		return nil, func(errGroup *jen.Statement, lock *jen.Statement) func(s *jen.Statement) {
			return func(s *jen.Statement) {
				if !g.doesFieldImplementsAppEnv(pkg, field) {
					return
				}

				localEnvsVariableName := "envs" + fieldName

				statement := jen.List(jen.Id(localEnvsVariableName), jen.Id("err")).Op(":=").Id("o").Dot(fieldName).Dot(methodeName).Call(jen.Id("ctx"))

				appendStatement := variable.Clone().Op("=").Append(
					variable,
					jen.Id(localEnvsVariableName).Op("..."),
				)

				if errGroup != nil {
					s.Add(jen.Add(errGroup).Dot("Go").Call(jen.Func().Params().Id("error").Block(
						statement,
						jen.If(jen.Id("err").Op("!=").Nil()).Block(jen.Return(jen.Id("err"))).Line(),
						lock.Line(),
						appendStatement,
						jen.Line().Return(jen.Nil()),
					)))
				} else {
					s.Add(statement).Line().
						If(jen.Id("err").Op("!=").Nil()).Block(jen.Return(jen.Nil(), jen.Id("err"))).Line().Line().
						Add(appendStatement)
				}
			}
		}, true, nil
	default:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			return nil, nil, false, errors.Errorf("type not yet supported: %v", ftyped)
		}

		embedded := field.Markers.Get(appenvmarkers.EmbeddedEnvironmentVariable)
		if embedded != nil {
			return nil, nil, false, errors.Errorf("type not yet supported: %v", ftyped)
		}

		return nil, nil, false, nil
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
