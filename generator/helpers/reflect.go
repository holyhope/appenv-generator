package helpers

import (
	"go/types"

	"sigs.k8s.io/controller-tools/pkg/loader"
)

const (
	MethodName = "GetApplicationEnvironments"
)

func HasGetApplicationEnvironmentsMethod(pkg *loader.Package, typeInfo types.Type) bool {
	method, ind, _ := types.LookupFieldOrMethod(typeInfo, true /* check pointers too */, pkg.Types, MethodName)
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
