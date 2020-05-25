package tests

type IgnoreAllField struct {
	// +appenv:ignore=true
	Simple SimpleTest `json:"simple"`

	// +appenv:key=SIMPLE_STRING
	// +appenv:ignore=true
	SimpleString string `json:"simpleString"`
}

type Ignore struct {
	// +appenv:key=SIMPLE_STRING
	SimpleString string `json:"simpleString"`

	// +appenv:key=SIMPLE_FLOAT
	// +appenv:ignore=true
	SimpleFloat float64 `json:"simpleFloat"`
}
