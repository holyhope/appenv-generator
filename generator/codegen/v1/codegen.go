package codegen

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

import (
	"go/types"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"

	"github.com/holyhope/appenv-generator/generator/helpers"
	appenvmarkers "github.com/holyhope/appenv-generator/markers"
)

type CodeGen struct {
	types map[implementsKey]implementsValue
}

func NewCodeGen() *CodeGen {
	return &CodeGen{
		types: map[implementsKey]implementsValue{},
	}
}

func (g *CodeGen) GenerateCode(pkg *loader.Package, info *markers.TypeInfo) (func(*jen.Statement), error) {
	const resultVariableName = "result"
	const errorVariableName = "err"

	staticEnvsStatements := []jen.Code{}
	staticEnvsFromStatements := []jen.Code{}
	statementstoAppend := []func(*jen.Statement){}
	takeTimeCount := 0

	for _, field := range info.Fields {
		ignore := field.Markers.Get(appenvmarkers.EnvironmentIgnore)

		if ignore != nil {
			continue
		}

		staticEnvsStatement, staticEnvsFromStatement, postStatement, takeTime, err := g.GenerateCodeWithField(pkg, jen.Id(resultVariableName), jen.Id(errorVariableName), field)
		if err != nil {
			return nil, errors.Wrapf(err, "field %s", field.Name)
		}

		if takeTime {
			takeTimeCount++
		}

		if staticEnvsStatement != nil {
			staticEnvsStatements = append(staticEnvsStatements, staticEnvsStatement)
		}
		if staticEnvsFromStatement != nil {
			staticEnvsFromStatements = append(staticEnvsFromStatements, staticEnvsFromStatement)
		}
		if postStatement != nil {
			statementstoAppend = append(statementstoAppend, postStatement)
		}
	}

	if len(statementstoAppend) == 0 && len(staticEnvsStatements) == 0 && len(staticEnvsFromStatements) == 0 {
		g.setTypeImplementsAppEnv(pkg, info, doesNotImplements)
		return nil, nil
	}

	if takeTimeCount == 0 {
		g.setTypeImplementsAppEnv(pkg, info, doesImplements)
	} else {
		g.setTypeImplementsAppEnv(pkg, info, inheritImplementation)
	}

	if len(staticEnvsStatements) != 0 || len(staticEnvsFromStatements) != 0 {
		g.setTypeImplementsAppEnv(pkg, info, doesImplements)
	}

	return func(s *jen.Statement) {
		s.Func().
			Params(jen.Id("o").Op("*").Id(info.Name)).
			Id(helpers.MethodName).
			Params(jen.Id("ctx").Qual("context", "Context")).
			Params(jen.Id("finalResult").Qual("github.com/holyhope/appenv-generator/v1", "Result"), jen.Id(errorVariableName).Id("error")).
			BlockFunc(func(g *jen.Group) {
				g.If(jen.Id("o").Op("==").Nil()).Block(jen.Return())
				g.Line()

				g.Id(resultVariableName).Op(":=").Qual("github.com/holyhope/appenv-generator/v1", "NewResult").Call(
					jen.Index().Qual("k8s.io/api/core/v1", "EnvVar").Values(staticEnvsStatements...),
					jen.Index().Qual("k8s.io/api/core/v1", "EnvFromSource").Values(staticEnvsFromStatements...),
				)
				g.Id("finalResult").Op("=").Id(resultVariableName)
				g.Line()

				if takeTimeCount == 0 {
					for _, statement := range statementstoAppend {
						g.Do(statement)
						g.If(jen.Id(errorVariableName).Op("!=").Nil()).Block(
							jen.Return(),
						)
						g.Line()
					}

					g.Return()
					return
				}

				g.List(jen.Id("errGroup"), jen.Id("ctx")).Op(":=").Qual("golang.org/x/sync/errgroup", "WithContext").Call(jen.Id("ctx"))
				g.Line()

				for _, statement := range statementstoAppend {
					g.Id("errGroup").Dot("Go").Call(
						jen.Func().Params().Parens(jen.Id(errorVariableName).Id("error")).Block(
							jen.Do(statement),
							jen.Return(),
						),
					).Line()
				}

				g.Id(errorVariableName).Op("=").Id("errGroup").Dot("Wait").Call()
				g.Return()
			})
	}, nil
}

func (g *CodeGen) GenerateCodeWithField(pkg *loader.Package, resultVariable, errorVariable *jen.Statement, field markers.FieldInfo) (*jen.Statement, *jen.Statement, func(*jen.Statement), bool, error) {
	ftype := pkg.TypesInfo.TypeOf(field.RawField.Type)

	if ftype == types.Typ[types.Invalid] {
		return nil, nil, nil, false, nil
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
			fromKind := field.Markers.Get(appenvmarkers.FromKindEnvironmentVariable)
			if fromKind != nil {
				if ftyped.Kind() != types.String {
					if isPointer {
						return nil, nil, nil, false, errors.Errorf("expected type *string, not *%s", ftyped.Name())
					}

					return nil, nil, nil, false, errors.Errorf("expected type string, not %s", ftyped.Name())
				}

				fromField := field.Markers.Get(appenvmarkers.FromFieldEnvironmentVariable)
				if fromField == nil {
					return nil, nil, nil, false, errors.Errorf("marker %s not found", appenvmarkers.FromFieldEnvironmentVariable)
				}

				var values jen.Code
				switch fromKind {
				case "secret":
					if isPointer {
						// Ensure field is not nil before dereferencing
						return nil, nil, func(s *jen.Statement) {
							s.
								If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
								Block(resultVariable.Clone().Dot("AddEnvs").Call(
									jen.Qual("k8s.io/api/core/v1", "EnvVar").Values(jen.Dict{
										jen.Id("Name"): jen.Lit(envVarName),
										jen.Id("ValueFrom"): jen.Op("&").Qual("k8s.io/api/core/v1", "EnvVarSource").Values(jen.Dict{
											jen.Id("SecretKeyRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "SecretKeySelector").Values(jen.Dict{
												jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
													jen.Id("Name"): jen.Op("*").Id("o").Dot(field.Name),
												}),
												jen.Id("Key"): jen.Lit(fromField),
											}),
										}),
									}),
								))
						}, false, nil
					}

					values = jen.Dict{
						jen.Id("SecretKeyRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "SecretKeySelector").Values(jen.Dict{
							jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
								jen.Id("Name"): jen.Id("o").Dot(field.Name),
							}),
							jen.Id("Key"): jen.Lit(fromField),
						}),
					}
				case "configMap":
					if isPointer {
						// Ensure field is not nil before dereferencing
						return nil, nil, func(s *jen.Statement) {
							s.
								If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
								Block(resultVariable.Clone().Dot("AddEnvs").Call(
									jen.Qual("k8s.io/api/core/v1", "EnvVar").Values(jen.Dict{
										jen.Id("Name"): jen.Lit(envVarName),
										jen.Id("ValueFrom"): jen.Op("&").Qual("k8s.io/api/core/v1", "EnvVarSource").Values(jen.Dict{
											jen.Id("ConfigMapKeyRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "ConfigMapKeySelector").Values(jen.Dict{
												jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
													jen.Id("Name"): jen.Op("*").Id("o").Dot(field.Name),
												}),
												jen.Id("Key"): jen.Lit(fromField),
											}),
										}),
									}),
								))
						}, false, nil
					}

					values = jen.Dict{
						jen.Id("ConfigMapKeyRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "ConfigMapKeySelector").Values(jen.Dict{
							jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
								jen.Id("Name"): jen.Id("o").Dot(field.Name),
							}),
							jen.Id("Key"): jen.Lit(fromField),
						}),
					}
				default:
					return nil, nil, nil, false, errors.Errorf("marker %s=%v not supported", appenvmarkers.FromKindEnvironmentVariable, fromKind)
				}

				return jen.Values(jen.Dict{
					jen.Id("Name"):      jen.Lit(envVarName),
					jen.Id("ValueFrom"): jen.Op("&").Qual("k8s.io/api/core/v1", "EnvVarSource").Values(values),
				}), nil, nil, false, nil
			}

			value := jen.Empty()
			if isPointer {
				// Ensure field is not nil before dereferencing
				return nil, nil, func(s *jen.Statement) {
					s.
						If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
						Block(resultVariable.Clone().Dot("AddEnvs").Call(
							jen.Qual("k8s.io/api/core/v1", "EnvVar").Values(jen.Dict{
								jen.Id("Name"):  jen.Lit(envVarName),
								jen.Id("Value"): value.Add(jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Op("*").Id("o").Dot(field.Name))),
							}),
						))
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
			}), nil, nil, false, nil
		}

		fromKind := field.Markers.Get(appenvmarkers.FromKindEnvironmentVariable)
		if fromKind != nil {
			fromField := field.Markers.Get(appenvmarkers.FromFieldEnvironmentVariable)
			if fromField != nil {
				return nil, nil, nil, false, errors.Errorf("unexpected marker %s", appenvmarkers.FromFieldEnvironmentVariable)
			}

			switch fromKind {
			case "secret":
				if isPointer {
					// Ensure field is not nil before dereferencing
					return nil, nil, func(s *jen.Statement) {
						s.
							If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
							Block(resultVariable.Clone().Dot("AddEnvsFrom").Call(
								jen.Qual("k8s.io/api/core/v1", "EnvFromSource").Values(jen.Dict{
									jen.Id("SecretRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "SecretEnvSource").Values(jen.Dict{
										jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
											jen.Id("Name"): jen.Op("*").Id("o").Dot(field.Name),
										}),
									}),
								}),
							))
					}, false, nil
				}

				return nil, jen.Values(jen.Dict{
					jen.Id("SecretRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "SecretEnvSource").Values(jen.Dict{
						jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
							jen.Id("Name"): jen.Id("o").Dot(field.Name),
						}),
					}),
				}), nil, false, nil
			case "configMap":
				if isPointer {
					// Ensure field is not nil before dereferencing
					return nil, nil, func(s *jen.Statement) {
						s.
							If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
							Block(resultVariable.Clone().Dot("AddEnvsFrom").Call(
								jen.Qual("k8s.io/api/core/v1", "EnvFromSource").Values(jen.Dict{
									jen.Id("ConfigMapRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "ConfigMapEnvSource").Values(jen.Dict{
										jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
											jen.Id("Name"): jen.Op("*").Id("o").Dot(field.Name),
										}),
									}),
								}),
							))
					}, false, nil
				}

				return nil, jen.Values(jen.Dict{
					jen.Id("ConfigMapRef"): jen.Op("&").Qual("k8s.io/api/core/v1", "ConfigMapEnvSource").Values(jen.Dict{
						jen.Id("LocalObjectReference"): jen.Qual("k8s.io/api/core/v1", "LocalObjectReference").Values(jen.Dict{
							jen.Id("Name"): jen.Id("o").Dot(field.Name),
						}),
					}),
				}), nil, false, nil
			default:
				return nil, nil, nil, false, errors.Errorf("marker %s=%v not supported", appenvmarkers.FromKindEnvironmentVariable, fromKind)
			}

		}

		return nil, nil, nil, false, nil

	case *types.Struct, *types.Interface, *types.Named:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			return jen.Values(jen.Dict{
				jen.Id("Name"):  jen.Lit(envVarName),
				jen.Id("Value"): jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Id("o").Dot(field.Name)),
			}), nil, nil, false, nil
		}

		fieldName := field.Name

		if fieldName == "" {
			fullNameParts := strings.Split(ftype.String(), ".")
			fieldName = fullNameParts[len(fullNameParts)-1]
		}

		//embedded := field.Markers.Get(appenvmarkers.EmbeddedEnvironmentVariable)
		//if embedded != nil {
		return nil, nil, func(s *jen.Statement) {
			if !g.FieldImplementsAppEnv(pkg, field) {
				return
			}

			localResultVariableName := "result" + fieldName
			localErrorVariableName := "err" + fieldName

			s.Block(
				jen.List(jen.Id(localResultVariableName), jen.Id(localErrorVariableName)).Op(":=").
					Id("o").Dot(fieldName).Dot(helpers.MethodName).Call(jen.Id("ctx")),
				jen.If(jen.Id(localErrorVariableName).Op("!=").Nil()).Block(
					errorVariable.Clone().Op("=").Id(localErrorVariableName),
					jen.Return(),
				),
				jen.Add(errorVariable).Op("=").Add(resultVariable).Dot("Add").Call(
					jen.Id("ctx"),
					jen.Id(localResultVariableName),
				),
			)
		}, true, nil
	default:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			return nil, nil, nil, false, errors.Errorf("type not yet supported: %v", ftyped)
		}

		return nil, nil, nil, false, nil
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
