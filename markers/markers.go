package appenv

const (
	GeneratorName = "appenv"
)

const (
	EnvironmentVariableName = GeneratorName + ":key"
	Ignore                  = GeneratorName + ":ignore"
)

const (
	FromKind      = GeneratorName + ":from:kind"
	FromFieldName = GeneratorName + ":from:field"
)

const (
	MountPath  = GeneratorName + ":mount:path"
	MountName  = GeneratorName + ":mount:name"
	MountKind  = GeneratorName + ":mount:kind"
	MountItems = GeneratorName + ":mount:items"
)
