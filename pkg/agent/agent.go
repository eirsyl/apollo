package agent

import (
	"errors"
	"fmt"

	"github.com/eirsyl/apollo/pkg"
	"github.com/eirsyl/apollo/pkg/agent/redis"
	"github.com/eirsyl/apollo/pkg/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Agent exports the proxy struct
type Agent struct {
	httpServer *HTTPServer
	client     *redis.Client
	executor   *Executor
}

// NewAgent initializes a new agent and returns a pointer to the instance
func NewAgent() (*Agent, error) {
	redisAddr := viper.GetString("redis")
	if redisAddr == "" {
		return nil, errors.New("The redis address cannot be empty")
	}

	managerAddr := viper.GetString("manager")
	if managerAddr == "" {
		return nil, errors.New("The manager address cannot be empty")
	}

	client, err := redis.NewClient(redisAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not create redis client: %v", err)
	}

	_, redisPort := utils.GetHostPort(redisAddr)
	httpPort := fmt.Sprintf(":%d", pkg.HTTPPortWindow+redisPort)
	httpServer, err := NewHTTPServer(httpPort)
	if err != nil {
		return nil, err
	}

	grpcPort := fmt.Sprintf(":%d", pkg.GRPCPortWindow+redisPort)
	executorServer, err := NewExecutor(grpcPort, managerAddr, client)
	if err != nil {
		return nil, err
	}

	return &Agent{
		client:     client,
		httpServer: httpServer,
		executor:   executorServer,
	}, nil
}

// Run the agent main functionality
func (a *Agent) Run() error {
	/*
	* Start the agent service
	*
	* Features:
	* - Monitor redis metrics
	* - Report cluster changes to the manager
	*
	 */
	var errChan = make(chan error, 1)

	go func(errChan chan error) {
		log.Infof("Starting http server on %s", a.httpServer.GetListenAddr())
		errChan <- a.httpServer.Run()
	}(errChan)

	go func(errChan chan error) {
		log.Infof("Starting executor on %s", a.executor.GetListenAddr())
		errChan <- a.executor.Run()
	}(errChan)

	return <-errChan
}

// Exit func
func (a *Agent) Exit() error {
	err := a.httpServer.Shutdown()
	if err != nil {
		return err
	}

	err = a.executor.Shutdown()
	if err != nil {
		return err
	}

	return a.client.Shutdown()
}