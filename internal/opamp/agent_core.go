package opamp

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server/types"
)

// Agent represents a connected Agent.
type Agent struct {
	// Some fields in this struct are exported so that we can render them in the UI.

	// Agent's instance id. This is an immutable field.
	InstanceId    uuid.UUID
	InstanceIdStr string

	// Group information
	GroupID   *string
	GroupName *string

	// Connection to the Agent.
	conn types.Connection

	// mutex for the fields that follow it.
	mux sync.RWMutex

	// Agent's current status.
	Status *protobufs.AgentToServer

	// The time when the agent has started. Valid only if Status.Health.Up==true
	StartedAt time.Time

	// Effective config reported by the Agent.
	EffectiveConfig string

	// Optional special remote config for this particular instance defined by
	// the user in the UI.
	CustomInstanceConfig string

	// Client certificate
	ClientCert                  *x509.Certificate
	ClientCertSha256Fingerprint string
	ClientCertOfferError        string

	// Remote config that we will give to this Agent.
	remoteConfig *protobufs.AgentRemoteConfig

	// Channels to notify when this Agent's status is updated next time.
	statusUpdateWatchers []chan<- struct{}
}

// NewAgent creates a new Agent instance
func NewAgent(
	instanceId uuid.UUID,
	conn types.Connection,
) *Agent {
	agent := &Agent{
		InstanceId:    instanceId,
		InstanceIdStr: instanceId.String(),
		conn:          conn,
	}
	tslConn, ok := conn.Connection().(*tls.Conn)
	if ok {
		// Client is using TLS connection.
		connState := tslConn.ConnectionState()
		if len(connState.PeerCertificates) > 0 {
			// Client uses client-side certificate. Get certificate details to display in the UI.
			leafClientCert := connState.PeerCertificates[0]
			fingerprint := sha256.Sum256(leafClientCert.Raw)
			agent.ClientCert = leafClientCert
			agent.ClientCertSha256Fingerprint = fmt.Sprintf("%X", fingerprint)
		}
	}

	return agent
}

// CloneReadonly returns a copy of the Agent that is safe to read.
// Functions that modify the Agent should not be called on the cloned copy.
func (agent *Agent) CloneReadonly() *Agent {
	agent.mux.RLock()
	defer agent.mux.RUnlock()

	return &Agent{
		InstanceId:                  agent.InstanceId,
		InstanceIdStr:               agent.InstanceIdStr,
		GroupID:                     agent.GroupID,
		GroupName:                   agent.GroupName,
		conn:                        agent.conn,
		Status:                      agent.Status,
		StartedAt:                   agent.StartedAt,
		EffectiveConfig:             agent.EffectiveConfig,
		CustomInstanceConfig:        agent.CustomInstanceConfig,
		ClientCert:                  agent.ClientCert,
		ClientCertSha256Fingerprint: agent.ClientCertSha256Fingerprint,
		ClientCertOfferError:        agent.ClientCertOfferError,
		remoteConfig:                agent.remoteConfig,
	}
}

// hasCapability checks if the agent has a specific capability
func (agent *Agent) hasCapability(capability protobufs.AgentCapabilities) bool {
	return agent.Status != nil && agent.Status.Capabilities&uint64(capability) != 0
}

// GetConnection returns the agent's connection
func (agent *Agent) GetConnection() types.Connection {
	return agent.conn
}

// GetRemoteConfig returns the agent's remote config (for internal use)
func (agent *Agent) GetRemoteConfig() *protobufs.AgentRemoteConfig {
	agent.mux.RLock()
	defer agent.mux.RUnlock()
	return agent.remoteConfig
}

// SetRemoteConfig sets the agent's remote config (for internal use)
func (agent *Agent) SetRemoteConfig(config *protobufs.AgentRemoteConfig) {
	agent.mux.Lock()
	defer agent.mux.Unlock()
	agent.remoteConfig = config
}

// AddStatusUpdateWatcher adds a channel to be notified when the agent's status updates
func (agent *Agent) AddStatusUpdateWatcher(ch chan<- struct{}) {
	agent.mux.Lock()
	defer agent.mux.Unlock()
	agent.statusUpdateWatchers = append(agent.statusUpdateWatchers, ch)
}

