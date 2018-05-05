package test_cases

type SlotCoverage struct {
}

func NewSlotCoverage() (*SlotCoverage, error) {
	return &SlotCoverage{}, nil
}

func (wd *SlotCoverage) GetName() string {
	return "slot-coverage"
}

func (wd *SlotCoverage) Run(args []string) error {
	return nil
}
