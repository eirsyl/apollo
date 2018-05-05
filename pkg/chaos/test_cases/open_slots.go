package test_cases

type OpenSlots struct {
}

func NewOpenSlots() (*OpenSlots, error) {
	return &OpenSlots{}, nil
}

func (wd *OpenSlots) GetName() string {
	return "open-slots"
}

func (wd *OpenSlots) Run(args []string) error {
	return nil
}
