package generator

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const (
	methodeName = "GetApplicationEnvironments"
)

func HasGetApplicationEnvironmentsMethod(pkg *loader.Package, typeInfo types.Type) bool {
	method, ind, _ := types.LookupFieldOrMethod(typeInfo, true /* check pointers too */, pkg.Types, methodeName)
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

func (g Generator) getKey(pkg *loader.Package, t types.Type) implementsKey {
	key := t.String()
	return implementsKey(strings.TrimLeft(key, "*"))
}

func (g Generator) getKeyForType(pkg *loader.Package, info *markers.TypeInfo) implementsKey {
	key := fmt.Sprintf("%s.%s", pkg.Package.ID, info.Name)
	return implementsKey(key)
}

func (g Generator) getKeyForField(pkg *loader.Package, info markers.FieldInfo) implementsKey {
	return g.getKey(pkg, pkg.TypesInfo.TypeOf(info.RawField.Type))
}

func (g Generator) setTypeImplementsAppEnv(pkg *loader.Package, info *markers.TypeInfo, value implements) {
	g.types[g.getKeyForType(pkg, info)] = implementsValue{
		Implements: value,
		Marker:     info,
	}
}

func (g Generator) doesTypeImplementsAppEnv(pkg *loader.Package, info *markers.TypeInfo) bool {
	value, ok := g.types[g.getKeyForType(pkg, info)]
	if !ok {
		return HasGetApplicationEnvironmentsMethod(pkg, pkg.TypesInfo.TypeOf(value.Marker.RawSpec.Type))
	}

	return g.doesValueImplements(pkg, value)
}

func (g Generator) doesFieldImplementsAppEnv(pkg *loader.Package, info markers.FieldInfo) bool {
	value, ok := g.types[g.getKeyForField(pkg, info)]
	if !ok {
		ftype := pkg.TypesInfo.TypeOf(info.RawField.Type)
		if HasGetApplicationEnvironmentsMethod(pkg, ftype) {
			return true
		}

		if ftype, ok := ftype.(*types.Pointer); ok {
			ftype := ftype.Elem()
			if _, ok := ftype.(*types.Pointer); ok {
				return HasGetApplicationEnvironmentsMethod(pkg, ftype)
			}

			value, ok := g.types[g.getKey(pkg, ftype)]
			if !ok {
				return HasGetApplicationEnvironmentsMethod(pkg, ftype)
			}

			return g.doesValueImplements(pkg, value)
		}

		return false
	}

	return g.doesTypeImplementsAppEnv(pkg, value.Marker)
}

func (g Generator) doesValueImplements(pkg *loader.Package, value implementsValue) bool {
	switch value.Implements {
	case doesImplements:
		return true

	case doesNotImplements:
		return HasGetApplicationEnvironmentsMethod(pkg, pkg.TypesInfo.TypeOf(value.Marker.RawSpec.Type))

	case inheritImplementation:
		for _, field := range value.Marker.Fields {
			if value, ok := g.types[g.getKeyForField(pkg, field)]; ok {
				if g.doesTypeImplementsAppEnv(pkg, value.Marker) {
					return true
				}
			}

			if HasGetApplicationEnvironmentsMethod(pkg, pkg.TypesInfo.TypeOf(field.RawField.Type)) {
				return true
			}
		}

		return false
	default:
		panic(errors.New("unexpected implements value"))
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
