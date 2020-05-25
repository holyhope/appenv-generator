package codegen

import (
	"errors"
	"fmt"
	"go/types"
	"strings"

	"github.com/holyhope/appenv-generator/generator/helpers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type implements string

const (
	doesImplements        implements = "true"
	doesNotImplements     implements = "false"
	inheritImplementation implements = "unknown"
)

type implementsKey string

type implementsValue struct {
	Implements implements
	Marker     *markers.TypeInfo
}

func (g *CodeGen) getKey(pkg *loader.Package, t types.Type) implementsKey {
	key := t.String()
	return implementsKey(strings.TrimLeft(key, "*"))
}

func (g *CodeGen) getKeyForType(pkg *loader.Package, info *markers.TypeInfo) implementsKey {
	key := fmt.Sprintf("%s.%s", pkg.Package.ID, info.Name)
	return implementsKey(key)
}

func (g *CodeGen) getKeyForField(pkg *loader.Package, info markers.FieldInfo) implementsKey {
	return g.getKey(pkg, pkg.TypesInfo.TypeOf(info.RawField.Type))
}

func (g *CodeGen) setTypeImplementsAppEnv(pkg *loader.Package, info *markers.TypeInfo, value implements) {
	g.types[g.getKeyForType(pkg, info)] = implementsValue{
		Implements: value,
		Marker:     info,
	}
}

func (g *CodeGen) TypeImplementsAppEnv(pkg *loader.Package, info *markers.TypeInfo) bool {
	value, ok := g.types[g.getKeyForType(pkg, info)]
	if !ok {
		return helpers.HasGetApplicationEnvironmentsMethod(pkg, pkg.TypesInfo.TypeOf(value.Marker.RawSpec.Type))
	}

	return g.doesValueImplements(pkg, value)
}

func (g *CodeGen) FieldImplementsAppEnv(pkg *loader.Package, info markers.FieldInfo) bool {
	value, ok := g.types[g.getKeyForField(pkg, info)]
	if !ok {
		ftype := pkg.TypesInfo.TypeOf(info.RawField.Type)
		if helpers.HasGetApplicationEnvironmentsMethod(pkg, ftype) {
			return true
		}

		if ftype, ok := ftype.(*types.Pointer); ok {
			ftype := ftype.Elem()
			if _, ok := ftype.(*types.Pointer); ok {
				return helpers.HasGetApplicationEnvironmentsMethod(pkg, ftype)
			}

			value, ok := g.types[g.getKey(pkg, ftype)]
			if !ok {
				return helpers.HasGetApplicationEnvironmentsMethod(pkg, ftype)
			}

			return g.doesValueImplements(pkg, value)
		}

		return false
	}

	return g.TypeImplementsAppEnv(pkg, value.Marker)
}

func (g *CodeGen) doesValueImplements(pkg *loader.Package, value implementsValue) bool {
	switch value.Implements {
	case doesImplements:
		return true

	case doesNotImplements:
		return helpers.HasGetApplicationEnvironmentsMethod(pkg, pkg.TypesInfo.TypeOf(value.Marker.RawSpec.Type))

	case inheritImplementation:
		for _, field := range value.Marker.Fields {
			if value, ok := g.types[g.getKeyForField(pkg, field)]; ok {
				if g.TypeImplementsAppEnv(pkg, value.Marker) {
					return true
				}
			}

			if helpers.HasGetApplicationEnvironmentsMethod(pkg, pkg.TypesInfo.TypeOf(field.RawField.Type)) {
				return true
			}
		}

		return false
	default:
		panic(errors.New("unexpected implements value"))
	}
}
