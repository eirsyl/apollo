package chaos

import (
	"errors"

	"github.com/eirsyl/apollo/pkg/chaos/test_cases"
)

// Chaos exports the chaos struct used to operate an chaos.
type Chaos struct {
	testCases []test_cases.Test
}

// NewChaos initializes a new chaos instance and returns a pointer to it.
func NewChaos() (*Chaos, error) {
	addNode, err := test_cases.NewAddNode()
	if err != nil {
		return nil, err
	}
	openSlots, err := test_cases.NewOpenSlots()
	if err != nil {
		return nil, err
	}
	slotCoverage, err := test_cases.NewSlotCoverage()
	if err != nil {
		return nil, err
	}
	writeData, err := test_cases.NewWriteData()
	if err != nil {
		return nil, err
	}

	return &Chaos{
		testCases: []test_cases.Test{
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
