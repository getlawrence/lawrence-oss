package opamp

import (
	"sync"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/protobufshelpers"
	"github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"
)

type Agents struct {
	mux         sync.RWMutex
	agentsById  map[uuid.UUID]*Agent
	connections map[types.Connection]map[uuid.UUID]bool
	logger      *zap.Logger
}

// NewAgents creates a new Agents instance with dependency injection
func NewAgents(logger *zap.Logger) *Agents {
	return &Agents{
		agentsById:  make(map[uuid.UUID]*Agent),
		connections: make(map[types.Connection]map[uuid.UUID]bool),
		logger:      logger,
	}
}

// RemoveConnection removes the connection and all Agent instances associated with the
// connection.
func (agents *Agents) RemoveConnection(conn types.Connection) {
	agents.mux.Lock()
	defer agents.mux.Unlock()

	// Get the list of agents to remove
	agentsToRemove := agents.connections[conn]

	// Remove from connections map
	delete(agents.connections, conn)

	// Remove the agents
	for agentId := range agentsToRemove {
		delete(agents.agentsById, agentId)
	}
}

func (agents *Agents) SetCustomConfigForAgent(
	agentId uuid.UUID,
	config *protobufs.AgentConfigMap,
	notifyNextStatusUpdate chan<- struct{},
) {
	agent := agents.FindAgent(agentId)
	if agent != nil {
		agent.SetCustomConfig(config, notifyNextStatusUpdate)
	}
}

func isEqualAgentDescr(d1, d2 *protobufs.AgentDescription) bool {
	if d1 == d2 {
		return true
	}
	if d1 == nil || d2 == nil {
		return false
	}
	return isEqualAttrs(d1.IdentifyingAttributes, d2.IdentifyingAttributes) &&
		isEqualAttrs(d1.NonIdentifyingAttributes, d2.NonIdentifyingAttributes)
}

func isEqualAttrs(attrs1, attrs2 []*protobufs.KeyValue) bool {
	if len(attrs1) != len(attrs2) {
		return false
	}
	for i, a1 := range attrs1 {
		a2 := attrs2[i]
		if !protobufshelpers.IsEqualKeyValue(a1, a2) {
			return false
		}
	}
	return true
}

func (agents *Agents) FindAgent(agentId uuid.UUID) *Agent {
	agents.mux.RLock()
	defer agents.mux.RUnlock()
	return agents.agentsById[agentId]
}

func (agents *Agents) FindOrCreateAgent(agentId uuid.UUID, conn types.Connection) *Agent {
	agents.mux.Lock()
	defer agents.mux.Unlock()

	// Ensure the Agent is in the agentsById map.
	agent := agents.agentsById[agentId]
	if agent == nil {
		agent = NewAgent(agentId, conn)
		agents.agentsById[agentId] = agent

		// Ensure the Agent's instance id is associated with the connection.
		if agents.connections[conn] == nil {
			agents.connections[conn] = map[uuid.UUID]bool{}
		}
		agents.connections[conn][agentId] = true
	}

	return agent
}

func (agents *Agents) GetAgentReadonlyClone(agentId uuid.UUID) *Agent {
	agent := agents.FindAgent(agentId)
	if agent == nil {
		return nil
	}

	// Return a clone to allow safe access after returning.
	return agent.CloneReadonly()
}

func (agents *Agents) GetAllAgentsReadonlyClone() map[uuid.UUID]*Agent {
	agents.mux.RLock()

	// Clone the map first
	m := map[uuid.UUID]*Agent{}
	for id, agent := range agents.agentsById {
		m[id] = agent
	}
	agents.mux.RUnlock()

	// Clone agents in the map
	for id, agent := range m {
		// Return a clone to allow safe access after returning.
		m[id] = agent.CloneReadonly()
	}
	return m
}

func (a *Agents) OfferAgentConnectionSettings(
	id uuid.UUID,
	offers *protobufs.ConnectionSettingsOffers,
) {
	a.logger.Info("Begin rotate client certificate", zap.String("agentId", id.String()))

	a.mux.Lock()
	defer a.mux.Unlock()

	agent, ok := a.agentsById[id]
	if ok {
		agent.OfferConnectionSettings(offers)
		a.logger.Info("Client certificate offers sent", zap.String("agentId", id.String()))
	} else {
		a.logger.Warn("Agent not found", zap.String("agentId", id.String()))
	}
}
