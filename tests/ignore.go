package tests

type Ignore struct {
	// +appenv:key=SIMPLE_STRING
	// +appenv:ignore=true
	SimpleString string `json:"simpleString"`

	// +appenv:ignore=true
	SimpleFloat float64 `json:"simpleFloat"`
}
