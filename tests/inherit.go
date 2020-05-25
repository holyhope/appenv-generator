package tests

type Inherit struct {
	SimpleTest
	*SimpleTest2
}

type InheritFromField struct {
	FromField
}
