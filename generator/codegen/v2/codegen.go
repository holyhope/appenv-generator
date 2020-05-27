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

const (
	kubernetesCoreV1 = "k8s.io/api/core/v1"
)

type CodeGen struct {
	types map[implementsKey]implementsValue
}

func NewCodeGen() *CodeGen {
	return &CodeGen{
		types: map[implementsKey]implementsValue{},
	}
}

const (
	argPosEnvs         = 0
	argPosEnvsFrom     = 1
	argPosVolumes      = 2
	argPosVolumesMount = 3
)

func (g *CodeGen) GenerateCode(pkg *loader.Package, info *markers.TypeInfo) (func(*jen.Statement), error) {
	const resultVariableName = "result"
	const errorVariableName = "err"

	initStatements := [4][]jen.Code{}
	statementstoAppend := []func(*jen.Statement){}
	takeTimeCount := 0

	for _, field := range info.Fields {
		ignore := field.Markers.Get(appenvmarkers.Ignore)

		if ignore != nil {
			continue
		}

		statements, postStatement, takeTime, err := g.GenerateCodeWithField(pkg, jen.Id(resultVariableName), jen.Id(errorVariableName), field)
		if err != nil {
			return nil, errors.Wrapf(err, "field %s", field.Name)
		}

		if takeTime {
			takeTimeCount++
		}

		for i, statements := range statements {
			initStatements[i] = append(initStatements[i], statements...)
		}

		if postStatement != nil {
			statementstoAppend = append(statementstoAppend, postStatement)
		}
	}

	if len(statementstoAppend) == 0 {
		doesNotImplement := true
		for _, statements := range initStatements {
			if len(statements) > 0 {
				doesNotImplement = false
				break
			}
		}
		if doesNotImplement {
			g.setTypeImplementsAppEnv(pkg, info, doesNotImplements)
			return nil, nil
		}
	}

	if takeTimeCount == 0 {
		g.setTypeImplementsAppEnv(pkg, info, doesImplements)
	} else {
		g.setTypeImplementsAppEnv(pkg, info, inheritImplementation)

		for _, statements := range initStatements {
			if len(statements) > 0 {
				g.setTypeImplementsAppEnv(pkg, info, doesImplements)
				break
			}
		}
	}

	return func(s *jen.Statement) {
		s.Func().
			Params(jen.Id("o").Op("*").Id(info.Name)).
			Id(helpers.MethodName).
			Params(jen.Id("ctx").Qual("context", "Context")).
			Params(jen.Id("finalResult").Qual("github.com/holyhope/appenv-generator/v2", "Result"), jen.Id(errorVariableName).Id("error")).
			BlockFunc(func(g *jen.Group) {
				g.If(jen.Id("o").Op("==").Nil()).Block(jen.Return())
				g.Line()

				initResult := jen.Qual("github.com/holyhope/appenv-generator/v2", "NewResult").Call(
					jen.Index().Qual(kubernetesCoreV1, "EnvVar").Values(initStatements[argPosEnvs]...),
					jen.Index().Qual(kubernetesCoreV1, "EnvFromSource").Values(initStatements[argPosEnvsFrom]...),
					jen.Index().Qual(kubernetesCoreV1, "Volume").Values(initStatements[argPosVolumes]...),
					jen.Index().Qual(kubernetesCoreV1, "VolumeMount").Values(initStatements[argPosVolumesMount]...),
				)

				if len(statementstoAppend) == 0 {
					g.Return(initResult, jen.Nil())
					return
				}

				g.Id(resultVariableName).Op(":=").Add(initResult)
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

func (g *CodeGen) GenerateCodeWithField(pkg *loader.Package, resultVariable, errorVariable *jen.Statement, field markers.FieldInfo) ([4][]jen.Code, func(*jen.Statement), bool, error) {
	ftype := pkg.TypesInfo.TypeOf(field.RawField.Type)

	if ftype == types.Typ[types.Invalid] {
		return [4][]jen.Code{nil, nil, nil, nil}, nil, false, nil
	}

	isPointer := false
	if ftyped, ok := ftype.(*types.Pointer); ok {
		ftype = ftyped.Elem()
		isPointer = true
	}

	switch ftyped := ftype.(type) {
	case *types.Basic:
		mountName := field.Markers.Get(appenvmarkers.MountName)
		if mountName != nil {
			if ftyped.Kind() != types.String {
				if isPointer {
					return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("expected type *string, not *%s", ftyped.Name())
				}

				return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("expected type string, not %s", ftyped.Name())
			}

			mountPath := field.Markers.Get(appenvmarkers.MountPath)
			if mountPath == nil {
				return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("marker %s not found", appenvmarkers.MountPath)
			}

			mountKind := field.Markers.Get(appenvmarkers.MountKind)
			if mountKind == nil {
				return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("marker %s not found", appenvmarkers.MountKind)
			}

			mountItems := field.Markers.Get(appenvmarkers.MountItems)

			var values [4][]jen.Code
			switch mountKind {
			case "secret":
				if isPointer {
					// Ensure field is not nil before dereferencing
					return [4][]jen.Code{nil, nil, nil, nil}, func(s *jen.Statement) {
						s.
							If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
							Block(resultVariable.Clone().Dot("AddEnvs").Call(
								jen.Qual(kubernetesCoreV1, "EnvVar").Values(jen.Dict{
									// ...
								}),
							))
					}, false, nil
				}

				source := jen.Dict{
					jen.Id("SecretName"): jen.Id("o").Dot(field.Name),
				}
				if mountItems != nil {
					source[jen.Id("Items")] = jen.Index().Qual(kubernetesCoreV1, "KeyToPath").ValuesFunc(func(g *jen.Group) {
						for key, value := range mountItems.(map[string]string) {
							g.Values(jen.Dict{
								jen.Id("Key"):  jen.Lit(key),
								jen.Id("Path"): jen.Lit(value),
							})
						}
					})
				}

				envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
				if envVarName != nil {
					values[argPosEnvs] = []jen.Code{
						jen.Values(jen.Dict{
							jen.Id("Name"):  jen.Lit(envVarName),
							jen.Id("Value"): jen.Lit(mountPath),
						}),
					}
				}

				values[argPosVolumes] = []jen.Code{
					jen.Values(jen.Dict{
						jen.Id("Name"): jen.Lit(mountName),
						jen.Id("VolumeSource"): jen.Qual(kubernetesCoreV1, "VolumeSource").Values(jen.Dict{
							jen.Id("Secret"): jen.Op("&").Qual(kubernetesCoreV1, "SecretVolumeSource").Values(source),
						}),
					}),
				}
				values[argPosVolumesMount] = []jen.Code{
					jen.Values(jen.Dict{
						jen.Id("Name"):      jen.Lit(mountName),
						jen.Id("MountPath"): jen.Lit(mountPath),
					}),
				}
			default:
				return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("marker %s=%v not supported", appenvmarkers.MountKind, mountKind)
			}

			return values, nil, false, nil
		}

		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			fromKind := field.Markers.Get(appenvmarkers.FromKind)
			if fromKind != nil {
				if ftyped.Kind() != types.String {
					if isPointer {
						return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("expected type *string, not *%s", ftyped.Name())
					}

					return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("expected type string, not %s", ftyped.Name())
				}

				fromField := field.Markers.Get(appenvmarkers.FromFieldName)
				if fromField == nil {
					return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("marker %s not found", appenvmarkers.FromFieldName)
				}

				var values jen.Code
				switch fromKind {
				case "secret":
					if isPointer {
						// Ensure field is not nil before dereferencing
						return [4][]jen.Code{nil, nil, nil, nil}, func(s *jen.Statement) {
							s.
								If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
								Block(resultVariable.Clone().Dot("AddEnvs").Call(
									jen.Qual(kubernetesCoreV1, "EnvVar").Values(jen.Dict{
										jen.Id("Name"): jen.Lit(envVarName),
										jen.Id("ValueFrom"): jen.Op("&").Qual(kubernetesCoreV1, "EnvVarSource").Values(jen.Dict{
											jen.Id("SecretKeyRef"): jen.Op("&").Qual(kubernetesCoreV1, "SecretKeySelector").Values(jen.Dict{
												jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
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
						jen.Id("SecretKeyRef"): jen.Op("&").Qual(kubernetesCoreV1, "SecretKeySelector").Values(jen.Dict{
							jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
								jen.Id("Name"): jen.Id("o").Dot(field.Name),
							}),
							jen.Id("Key"): jen.Lit(fromField),
						}),
					}
				case "configMap":
					if isPointer {
						// Ensure field is not nil before dereferencing
						return [4][]jen.Code{nil, nil, nil, nil}, func(s *jen.Statement) {
							s.
								If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
								Block(resultVariable.Clone().Dot("AddEnvs").Call(
									jen.Qual(kubernetesCoreV1, "EnvVar").Values(jen.Dict{
										jen.Id("Name"): jen.Lit(envVarName),
										jen.Id("ValueFrom"): jen.Op("&").Qual(kubernetesCoreV1, "EnvVarSource").Values(jen.Dict{
											jen.Id("ConfigMapKeyRef"): jen.Op("&").Qual(kubernetesCoreV1, "ConfigMapKeySelector").Values(jen.Dict{
												jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
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
						jen.Id("ConfigMapKeyRef"): jen.Op("&").Qual(kubernetesCoreV1, "ConfigMapKeySelector").Values(jen.Dict{
							jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
								jen.Id("Name"): jen.Id("o").Dot(field.Name),
							}),
							jen.Id("Key"): jen.Lit(fromField),
						}),
					}
				default:
					return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("marker %s=%v not supported", appenvmarkers.FromKind, fromKind)
				}

				return [4][]jen.Code{{jen.Values(jen.Dict{
					jen.Id("Name"):      jen.Lit(envVarName),
					jen.Id("ValueFrom"): jen.Op("&").Qual(kubernetesCoreV1, "EnvVarSource").Values(values),
				})}, nil, nil, nil}, nil, false, nil
			}

			value := jen.Empty()
			if isPointer {
				// Ensure field is not nil before dereferencing
				return [4][]jen.Code{nil, nil, nil, nil}, func(s *jen.Statement) {
					s.
						If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
						Block(resultVariable.Clone().Dot("AddEnvs").Call(
							jen.Qual(kubernetesCoreV1, "EnvVar").Values(jen.Dict{
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

			return [4][]jen.Code{{jen.Values(jen.Dict{
				jen.Id("Name"):  jen.Lit(envVarName),
				jen.Id("Value"): value,
			})}, nil, nil, nil}, nil, false, nil
		}

		fromKind := field.Markers.Get(appenvmarkers.FromKind)
		if fromKind != nil {
			fromField := field.Markers.Get(appenvmarkers.FromFieldName)
			if fromField != nil {
				return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("unexpected marker %s", appenvmarkers.FromFieldName)
			}

			switch fromKind {
			case "secret":
				if isPointer {
					// Ensure field is not nil before dereferencing
					return [4][]jen.Code{nil, nil, nil, nil}, func(s *jen.Statement) {
						s.
							If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
							Block(resultVariable.Clone().Dot("AddEnvsFrom").Call(
								jen.Qual(kubernetesCoreV1, "EnvFromSource").Values(jen.Dict{
									jen.Id("SecretRef"): jen.Op("&").Qual(kubernetesCoreV1, "SecretEnvSource").Values(jen.Dict{
										jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
											jen.Id("Name"): jen.Op("*").Id("o").Dot(field.Name),
										}),
									}),
								}),
							))
					}, false, nil
				}

				return [4][]jen.Code{nil, {jen.Values(jen.Dict{
					jen.Id("SecretRef"): jen.Op("&").Qual(kubernetesCoreV1, "SecretEnvSource").Values(jen.Dict{
						jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
							jen.Id("Name"): jen.Id("o").Dot(field.Name),
						}),
					}),
				})}, nil, nil}, nil, false, nil
			case "configMap":
				if isPointer {
					// Ensure field is not nil before dereferencing
					return [4][]jen.Code{nil, nil, nil, nil}, func(s *jen.Statement) {
						s.
							If(jen.Id("o").Dot(field.Name).Op("!=").Nil()).
							Block(resultVariable.Clone().Dot("AddEnvsFrom").Call(
								jen.Qual(kubernetesCoreV1, "EnvFromSource").Values(jen.Dict{
									jen.Id("ConfigMapRef"): jen.Op("&").Qual(kubernetesCoreV1, "ConfigMapEnvSource").Values(jen.Dict{
										jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
											jen.Id("Name"): jen.Op("*").Id("o").Dot(field.Name),
										}),
									}),
								}),
							))
					}, false, nil
				}

				return [4][]jen.Code{nil, {jen.Values(jen.Dict{
					jen.Id("ConfigMapRef"): jen.Op("&").Qual(kubernetesCoreV1, "ConfigMapEnvSource").Values(jen.Dict{
						jen.Id("LocalObjectReference"): jen.Qual(kubernetesCoreV1, "LocalObjectReference").Values(jen.Dict{
							jen.Id("Name"): jen.Id("o").Dot(field.Name),
						}),
					}),
				})}, nil, nil}, nil, false, nil
			default:
				return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("marker %s=%v not supported", appenvmarkers.FromKind, fromKind)
			}

		}

		return [4][]jen.Code{nil, nil, nil, nil}, nil, false, nil

	case *types.Struct, *types.Interface, *types.Named:
		envVarName := field.Markers.Get(appenvmarkers.EnvironmentVariableName)
		if envVarName != nil {
			return [4][]jen.Code{{jen.Values(jen.Dict{
				jen.Id("Name"):  jen.Lit(envVarName),
				jen.Id("Value"): jen.Qual("fmt", "Sprintf").Call(jen.Lit("%v"), jen.Id("o").Dot(field.Name)),
			})}, nil, nil, nil}, nil, false, nil
		}

		fieldName := field.Name

		if fieldName == "" {
			fullNameParts := strings.Split(ftype.String(), ".")
			fieldName = fullNameParts[len(fullNameParts)-1]
		}

		//embedded := field.Markers.Get(appenvmarkers.EmbeddedEnvironmentVariable)
		//if embedded != nil {
		return [4][]jen.Code{nil, nil, nil, nil}, func(s *jen.Statement) {
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
			return [4][]jen.Code{nil, nil, nil, nil}, nil, false, errors.Errorf("type not yet supported: %v", ftyped)
		}

		return [4][]jen.Code{nil, nil, nil, nil}, nil, false, nil
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
