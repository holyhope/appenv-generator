package generator

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-tools/pkg/markers"

	appenvmarkers "github.com/holyhope/appenv-generator/markers"
)

var (
	keyFieldMarker    = markers.Must(markers.MakeDefinition(appenvmarkers.EnvironmentVariableName, markers.DescribesField, ""))
	ignoreFieldMarker = markers.Must(markers.MakeDefinition(appenvmarkers.Ignore, markers.DescribesField, true))
)
var (
	fromKindMarker  = markers.Must(markers.MakeDefinition(appenvmarkers.FromKind, markers.DescribesField, ""))
	fromFieldMarker = markers.Must(markers.MakeDefinition(appenvmarkers.FromFieldName, markers.DescribesField, ""))
)
var (
	mountKindMarker  = markers.Must(markers.MakeDefinition(appenvmarkers.MountKind, markers.DescribesField, ""))
	mountNameMarker  = markers.Must(markers.MakeDefinition(appenvmarkers.MountName, markers.DescribesField, ""))
	mountPathMarker  = markers.Must(markers.MakeDefinition(appenvmarkers.MountPath, markers.DescribesField, ""))
	mountItemsMarker = markers.Must(markers.MakeDefinition(appenvmarkers.MountItems, markers.DescribesField, map[string]string{}))
)

func (Generator) RegisterMarkers(into *markers.Registry) error {
	if err := markers.RegisterAll(into, ignoreFieldMarker); err != nil {
		return errors.Wrap(err, "cannot register basic markers")
	}

	if err := markers.RegisterAll(into, keyFieldMarker, fromFieldMarker, fromKindMarker); err != nil {
		return errors.Wrap(err, "cannot register env markers")
	}

	if err := markers.RegisterAll(into, mountKindMarker, mountNameMarker, mountPathMarker, mountItemsMarker); err != nil {
		return errors.Wrap(err, "cannot register mount markers")
	}

	into.AddHelp(
		keyFieldMarker,
		markers.SimpleHelp("object", "The name of the environment variable to inject into the container"),
	)

	return nil
}
