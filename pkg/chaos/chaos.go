package chaos

import (
	"errors"

	"github.com/eirsyl/apollo/pkg/chaos/testcases"
)

// Chaos exports the chaos struct used to operate an chaos.
type Chaos struct {
	testCases []testcases.Test
}

// NewChaos initializes a new chaos instance and returns a pointer to it.
func NewChaos() (*Chaos, error) {
	addNode, err := testcases.NewAddNode()
	if err != nil {
		return nil, err
	}
	openSlots, err := testcases.NewOpenSlots()
	if err != nil {
		return nil, err
	}
	slotCoverage, err := testcases.NewSlotCoverage()
	if err != nil {
		return nil, err
	}
	writeData, err := testcases.NewWriteData()
	if err != nil {
		return nil, err
	}

	return &Chaos{
		testCases: []testcases.Test{
			addNode,
			openSlots,
			slotCoverage,
			writeData,
		},
	}, nil
}

// Run starts the Chaos
func (m *Chaos) Run(operation string, args []string) error {
	for _, test := range m.testCases {
		name := test.GetName()
		if name == operation {
			return test.Run(args)
		}
	}
	return errors.New("invalid operation")
}

// Exit gracefully shuts down the chaos
func (m *Chaos) Exit() error {
	// Not implemented
	return nil
}
