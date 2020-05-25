package generator

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/holyhope/appenv-generator/generator/helpers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

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
		if helpers.HasGetApplicationEnvironmentsMethod(pkg, typeInfo) {
			return false
		}

		for underlyingType := typeInfo.Underlying(); underlyingType != lastType; lastType, underlyingType = underlyingType, underlyingType.Underlying() {
			// if it has a manual shallowcopy, we're fine
			if helpers.HasGetApplicationEnvironmentsMethod(pkg, underlyingType) {
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
