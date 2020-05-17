package assets

type SimpleTest struct {
	// +appenv:key=SIMPLE_STRING
	SimpleString string `json:"simpleString"`
}

type NoOp struct {
	// +appenv:ignore=true
	AField interface{} `json:"aField"`
}

type Test struct {
	// +appenv:key=THE_NUMBER
	TheNumber float64 `json:"theNumber"`

	// +appenv:ignore=true
	RecursiveField *Test `json:"recursive"`

	// +appenv:embedded=true
	SimpleField *SimpleTest `json:"simpleField"`

	// +appenv:key=SIMPLE_TEST_VALUE
	SimpleTest *SimpleTest `json:"simpleTest"`

	NoOp *NoOp `json:"noOp"`
}
