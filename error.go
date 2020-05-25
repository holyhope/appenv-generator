package appenv

import (
	"fmt"
)

type versionError struct {
	Cause Result
}

func NewVersionError(cause Result) error {
	return &versionError{cause}
}

func (e *versionError) Error() string {
	return fmt.Sprintf("unsupported version (%+v)", e.Cause)
}
