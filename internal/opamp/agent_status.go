package opamp

import (
	"bytes"
	"time"

	"github.com/open-telemetry/opamp-go/protobufs"
	"google.golang.org/protobuf/proto"
)

// UpdateStatus processes the OpAMP status message received from the Agent. Updates Agent's
// status report and sets appropriate fields in the response message to be sent
// to the Agent.
func (agent *Agent) UpdateStatus(
	statusMsg *protobufs.AgentToServer,
	response *protobufs.ServerToAgent,
) {
	agent.mux.Lock()

	agent.processStatusUpdate(statusMsg, response)

	if statusMsg.ConnectionSettingsRequest != nil {
		agent.processConnectionSettingsRequest(statusMsg.ConnectionSettingsRequest.Opamp, response)
	}

	statusUpdateWatchers := agent.statusUpdateWatchers
	agent.statusUpdateWatchers = nil

	agent.mux.Unlock()

	// Notify watcher outside mutex to avoid blocking the mutex for too long.
	notifyStatusWatchers(statusUpdateWatchers)
}

// notifyStatusWatchers notifies all watchers that status has been updated
func notifyStatusWatchers(statusUpdateWatchers []chan<- struct{}) {
	// Notify everyone who is waiting on this Agent's status updates.
	for _, ch := range statusUpdateWatchers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// updateAgentDescription updates the agent's description and returns whether it changed
func (agent *Agent) updateAgentDescription(newStatus *protobufs.AgentToServer) (agentDescrChanged bool) {
	prevStatus := agent.Status

	if agent.Status == nil {
		// First time this Agent reports a status, remember it.
		agent.Status = newStatus
		agentDescrChanged = true
	} else {
		// Not a new Agent. Update the Status.
		agent.Status.SequenceNum = newStatus.SequenceNum

		// Check what's changed in the AgentDescription.
		if newStatus.AgentDescription != nil {
			// If the AgentDescription field is set it means the Agent tells us
			// something is changed in the field since the last status report
			// (or this is the first report).
			// Make full comparison of previous and new descriptions to see if it
			// really is different.
			if prevStatus != nil && proto.Equal(prevStatus.AgentDescription, newStatus.AgentDescription) {
				// Agent description didn't change.
				agentDescrChanged = false
			} else {
				// Yes, the description is different, update it.
				agent.Status.AgentDescription = newStatus.AgentDescription
				agentDescrChanged = true
			}
		} else {
			// AgentDescription field is not set, which means description didn't change.
			agentDescrChanged = false
		}

		// Update remote config status if it is included and is different from what we have.
		if newStatus.RemoteConfigStatus != nil &&
			!proto.Equal(agent.Status.RemoteConfigStatus, newStatus.RemoteConfigStatus) {
			agent.Status.RemoteConfigStatus = newStatus.RemoteConfigStatus
		}
	}
	return agentDescrChanged
}

// updateHealth updates the agent's health status
func (agent *Agent) updateHealth(newStatus *protobufs.AgentToServer) {
	if newStatus.Health == nil {
		return
	}

	agent.Status.Health = newStatus.Health

	if agent.Status != nil && agent.Status.Health != nil && agent.Status.Health.Healthy {
		agent.StartedAt = time.Unix(0, int64(agent.Status.Health.StartTimeUnixNano)).UTC()
	}
}

// updateRemoteConfigStatus updates the remote config status
func (agent *Agent) updateRemoteConfigStatus(newStatus *protobufs.AgentToServer) {
	// Update remote config status if it is included and is different from what we have.
	if newStatus.RemoteConfigStatus != nil {
		agent.Status.RemoteConfigStatus = newStatus.RemoteConfigStatus
	}
}

// updateStatusField updates the status field and returns whether agent description changed
func (agent *Agent) updateStatusField(newStatus *protobufs.AgentToServer) (agentDescrChanged bool) {
	if agent.Status == nil {
		// First time this Agent reports a status, remember it.
		agent.Status = newStatus
		agentDescrChanged = true
	}

	agentDescrChanged = agent.updateAgentDescription(newStatus) || agentDescrChanged
	agent.updateHealth(newStatus)
	agent.updateRemoteConfigStatus(newStatus)

	return agentDescrChanged
}

// processStatusUpdate processes the status update message
func (agent *Agent) processStatusUpdate(
	newStatus *protobufs.AgentToServer,
	response *protobufs.ServerToAgent,
) {
	// We don't have any status for this Agent, or we lost the previous status update from the Agent, so our
	// current status is not up-to-date.
	lostPreviousUpdate := (agent.Status == nil) || (agent.Status != nil && agent.Status.SequenceNum+1 != newStatus.SequenceNum)

	agentDescrChanged := agent.updateStatusField(newStatus)

	// Check if any fields were omitted in the status report.
	effectiveConfigOmitted := newStatus.EffectiveConfig == nil &&
		agent.hasCapability(protobufs.AgentCapabilities_AgentCapabilities_ReportsEffectiveConfig)

	packageStatusesOmitted := newStatus.PackageStatuses == nil &&
		agent.hasCapability(protobufs.AgentCapabilities_AgentCapabilities_ReportsPackageStatuses)

	remoteConfigStatusOmitted := newStatus.RemoteConfigStatus == nil &&
		agent.hasCapability(protobufs.AgentCapabilities_AgentCapabilities_ReportsRemoteConfig)

	healthOmitted := newStatus.Health == nil &&
		agent.hasCapability(protobufs.AgentCapabilities_AgentCapabilities_ReportsHealth)

	// True if the status was not fully reported.
	statusIsCompressed := effectiveConfigOmitted || packageStatusesOmitted || remoteConfigStatusOmitted || healthOmitted

	if statusIsCompressed && lostPreviousUpdate {
		// The status message is not fully set in the message that we received, but we lost the previous
		// status update. Request full status update from the agent.
		response.Flags |= uint64(protobufs.ServerToAgentFlags_ServerToAgentFlags_ReportFullState)
	}

	configChanged := false
	if agentDescrChanged {
		// Agent description is changed.

		// We need to recalculate the config.
		configChanged = agent.calcRemoteConfig()

		// And set connection settings that are appropriate for the Agent description.
		agent.calcConnectionSettings(response)
	}

	// If remote config is changed and different from what the Agent has then
	// send the new remote config to the Agent.
	if configChanged ||
		(agent.Status.RemoteConfigStatus != nil &&
			!bytes.Equal(agent.Status.RemoteConfigStatus.LastRemoteConfigHash, agent.remoteConfig.ConfigHash)) {
		// The new status resulted in a change in the config of the Agent or the Agent
		// does not have this config (hash is different). Send the new config the Agent.
		response.RemoteConfig = agent.remoteConfig
	}

	agent.updateEffectiveConfig(newStatus, response)
}

// updateEffectiveConfig updates the effective config from the agent's status
func (agent *Agent) updateEffectiveConfig(
	newStatus *protobufs.AgentToServer,
	response *protobufs.ServerToAgent,
) {
	if newStatus.EffectiveConfig != nil {
		// Agent reports effective config. Store it.
		if len(newStatus.EffectiveConfig.ConfigMap.ConfigMap) > 0 {
			// Get the first config (agents typically send one config file)
			for _, configFile := range newStatus.EffectiveConfig.ConfigMap.ConfigMap {
				agent.EffectiveConfig = string(configFile.Body)
				break
			}
		}
	}
}
