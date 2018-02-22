package redis

import (
	"fmt"
)

// ErrNodeIncompatible error is returned by the prechecks if the node
// isn't ready for apollo
type ErrNodeIncompatible struct {
	details []string
}

// NewErrNodeIncompatible returns a new incompatible error
func NewErrNodeIncompatible(details []string) *ErrNodeIncompatible {
	return &ErrNodeIncompatible{details: details}
}

func (e *ErrNodeIncompatible) Error() string {
	return fmt.Sprintf("Node incompatible: %v", e.details)
}

// ScrapeResult contains node information collected from the redis instance
type ScrapeResult struct {
	Name  string
	Value float64
}
