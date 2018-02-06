package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
	redisPort  int
}

// NewAgent initializes a new agent and returns a pointer to the instance
func NewAgent() (*Agent, error) {
	redisAddr := viper.GetString("redis")
	if redisAddr == "" {
		return nil, errors.New("The redis address cannot be empty")
	}

	client, err := redis.NewClient(redisAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not create redis client: %v", err)
	}

	_, redisPort := utils.GetHostPort(redisAddr)

	return &Agent{
		client:    client,
		redisPort: redisPort,
	}, nil
}

// Run func
func (a *Agent) Run() error {
	/*
	* Start the agent service
	*
	* Features:
	* - Monitor redis metrics
	* - Report cluster changes to the manager
	*
	 */
	httpServer, err := NewHTTPServer(
		fmt.Sprintf(":%d", pkg.PortWindow+a.redisPort),
	)
	if err != nil {
		return err
	}
	a.httpServer = httpServer

	log.Infof("Starting http server on %s", a.httpServer.SRV.Addr)
	err = a.httpServer.SRV.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Exit func
func (a *Agent) Exit() error {
	err := a.httpServer.SRV.Shutdown(context.Background())
	if err != nil {
		return err
	}
	return a.client.Shutdown()
}
