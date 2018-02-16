package agent

import (
	"github.com/eirsyl/apollo/pkg/agent/redis"
)

// Executor implements an GRPC server and a reconciliation loop
// that it used for detecting problems with the redis instance.
type Executor struct {
	redis       *redis.Client
	managerAddr string
	loop        *ReconciliationLoop
}

// NewExecutor creates a new Executor instance
func NewExecutor(managerAddr string, redis *redis.Client) (*Executor, error) {
	loop, err := NewReconciliationLoop(redis, managerAddr)
	if err != nil {
		return nil, err
	}

	executor := &Executor{
		managerAddr: managerAddr,
		redis:       redis,
		loop:        loop,
	}

	return executor, nil
}

// Run starts the executor instance
func (e *Executor) Run() error {
	var errChan = make(chan error, 1)

	go func() {
		errChan <- e.loop.Run()
	}()

	return <-errChan
}

// Shutdown stops the executor gracefully
func (e *Executor) Shutdown() error {
	return e.loop.Shutdown()
}
