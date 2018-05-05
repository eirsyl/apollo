package test_cases

type Test interface {
	GetName() string
	Run(args []string) error
}
