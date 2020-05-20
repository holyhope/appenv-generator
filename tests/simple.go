package tests

type SimpleTest struct {
	// +appenv:key=SIMPLE_STRING
	SimpleString string `json:"simpleString"`

	// +appenv:key=SIMPLE_FLOAT
	SimpleFloat float64 `json:"simpleFloat"`

	// +appenv:key=SIMPLE_INT
	SimpleInt int8 `json:"simpleInt"`

	// +appenv:key=SIMPLE_BOOL
	SimpleBool bool `json:"simpleBool"`
}

type SimpleTest2 struct {
	// +appenv:key=SIMPLE_STRING2
	SimpleString2 *string `json:"simpleString2"`
}
