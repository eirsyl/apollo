package test_cases

type BalanceCluster struct {
}

func NewBalanceCluster() (*BalanceCluster, error) {
	return &BalanceCluster{}, nil
}

func (wd *BalanceCluster) GetName() string {
	return "balance-cluster"
}

func (wd *BalanceCluster) Run(args []string) error {
	return nil
}
