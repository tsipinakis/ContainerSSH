package docker

import (
	"net"
	"reflect"
	"sync"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/internal/structutils"
	log2 "github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
)

// New creates a new NetworkConnectionHandler for a specific client.
func New(
	client net.TCPAddr,
	connectionID string,
	cfg config.DockerConfig,
	logger log2.Logger,
	backendRequestsMetric metrics.SimpleCounter,
	backendFailuresMetric metrics.SimpleCounter,
) (
	sshserver.NetworkConnectionHandler,
	error,
) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if cfg.Execution.DisableAgent {
		logger.Warning(message.NewMessage(message.EDockerGuestAgentDisabled, "ContainerSSH Guest Agent support is disabled. Some functions will not work."))
		defaultCfg := &config.DockerConfig{}
		structutils.Defaults(defaultCfg)
		if cfg.Execution.Mode == config.DockerExecutionModeConnection && reflect.DeepEqual(cfg.Execution.IdleCommand, defaultCfg.Execution.IdleCommand) {
			logger.Warning(message.NewMessage(message.EDockerGuestAgentDisabled, "ContainerSSH Guest Agent support is disabled, but the execution mode is set to connection and the idle command still points to the guest agent to provide an init program. This is very likely to break since you most likely don't have the guest agent installed."))
		}
	}

	return &networkHandler{
		mutex:        &sync.Mutex{},
		client:       client,
		connectionID: connectionID,
		config:       cfg,
		logger:       logger,
		disconnected: false,
		dockerClientFactory: &dockerV20ClientFactory{
			backendFailuresMetric: backendFailuresMetric,
			backendRequestsMetric: backendRequestsMetric,
		},
		done: make(chan struct{}),
	}, nil
}