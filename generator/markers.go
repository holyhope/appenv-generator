package generator

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/markers"

	appenvmarkers "github.com/holyhope/appenv-generator/markers"
)

var (
	keyFieldMarker    = markers.Must(markers.MakeDefinition(appenvmarkers.EnvironmentVariableName, markers.DescribesField, ""))
	ignoreFieldMarker = markers.Must(markers.MakeDefinition(appenvmarkers.Ignore, markers.DescribesField, true))
	fromKindMarker    = markers.Must(markers.MakeDefinition(appenvmarkers.FromKind, markers.DescribesField, ""))
	fromFieldMarker   = markers.Must(markers.MakeDefinition(appenvmarkers.FromFieldName, markers.DescribesField, ""))
)

func (Generator) RegisterMarkers(into *markers.Registry) error {
	if err := markers.RegisterAll(into, keyFieldMarker, ignoreFieldMarker, fromFieldMarker, fromKindMarker); err != nil {
		return errors.Wrap(err, "cannot register markers")
	}

	into.AddHelp(
		keyFieldMarker,
		markers.SimpleHelp("object", "The name of the environment variable to inject into the container"),
	)

	return nil
}
