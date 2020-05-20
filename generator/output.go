package generator

// Inpsired by https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu
// and https://github.com/banzaicloud/go-code-generation-demo

import (
	"io"

	appenv "github.com/holyhope/appenv-generator/markers"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
)

const (
	outputFile = "zz_generated." + appenv.GeneratorName + ".go"
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
