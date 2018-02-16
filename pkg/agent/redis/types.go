package redis

import (
	"fmt"
)

// ErrInstanceIncompatible error is returned by the prechecks if the instance
// isn't ready for apollo
type ErrInstanceIncompatible struct {
	details []string
}

// NewErrInstanceIncompatible returns a new incompatible error
func NewErrInstanceIncompatible(details []string) *ErrInstanceIncompatible {
	return &ErrInstanceIncompatible{details: details}
}

func (e *ErrInstanceIncompatible) Error() string {
	return fmt.Sprintf("Instance incompatible: %v", e.details)
}

type scrapeResult struct {
	Name  string
	Value float64
}
