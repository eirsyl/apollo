package testcases

// Test represents the interface used to implement test cases
type Test interface {
	GetName() string
	Run(args []string) error
}
