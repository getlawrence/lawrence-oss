package opamp

import (
	"bytes"
	"crypto/sha256"

	"github.com/open-telemetry/opamp-go/protobufs"
)

// SetCustomConfig sets a custom config for this Agent.
// notifyWhenConfigIsApplied channel is notified after the remote config is applied
// to the Agent and after the Agent reports back the effective config.
// If the provided config is equal to the current remoteConfig of the Agent
// then we will not send any config to the Agent and notifyWhenConfigIsApplied channel
// will be notified immediately. This requires that notifyWhenConfigIsApplied channel
// has a buffer size of at least 1.
func (agent *Agent) SetCustomConfig(
	config *protobufs.AgentConfigMap,
	notifyWhenConfigIsApplied chan<- struct{},
) {
	agent.mux.Lock()

	agent.CustomInstanceConfig = string(config.ConfigMap[""].Body)

	configChanged := agent.calcRemoteConfig()
	if configChanged {
		if notifyWhenConfigIsApplied != nil {
			// The caller wants to be notified when the Agent reports a status
			// update next time. This is typically used in the UI to wait until
			// the configuration changes are propagated successfully to the Agent.
			agent.statusUpdateWatchers = append(
				agent.statusUpdateWatchers,
				notifyWhenConfigIsApplied,
			)
		}
		msg := &protobufs.ServerToAgent{
			RemoteConfig: agent.remoteConfig,
		}
		agent.mux.Unlock()

		agent.SendToAgent(msg)
	} else {
		agent.mux.Unlock()

		if notifyWhenConfigIsApplied != nil {
			// No config change. We are not going to send config to the Agent and
			// as a result we do not expect status update from the Agent, so we will
			// just notify the waiter that the config change is done.
			notifyWhenConfigIsApplied <- struct{}{}
		}
	}
}

// calcRemoteConfig calculates the remote config for this Agent. It returns true if
// the calculated new config is different from the existing config stored in
// Agent.remoteConfig.
func (agent *Agent) calcRemoteConfig() bool {
	hash := sha256.New()

	cfg := protobufs.AgentRemoteConfig{
		Config: &protobufs.AgentConfigMap{
			ConfigMap: map[string]*protobufs.AgentConfigFile{},
		},
	}

	// Add the custom config for this particular Agent instance. Use empty
	// string as the config file name.
	cfg.Config.ConfigMap[""] = &protobufs.AgentConfigFile{
		Body: []byte(agent.CustomInstanceConfig),
	}

	// Calculate the hash.
	for k, v := range cfg.Config.ConfigMap {
		hash.Write([]byte(k))
		hash.Write(v.Body)
		hash.Write([]byte(v.ContentType))
	}

	cfg.ConfigHash = hash.Sum(nil)

	configChanged := !isEqualRemoteConfig(agent.remoteConfig, &cfg)

	agent.remoteConfig = &cfg

	return configChanged
}

// Configuration comparison helper functions

// isEqualRemoteConfig compares two remote configurations
func isEqualRemoteConfig(c1, c2 *protobufs.AgentRemoteConfig) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	return isEqualConfigSet(c1.Config, c2.Config)
}

// isEqualConfigSet compares two configuration sets
func isEqualConfigSet(c1, c2 *protobufs.AgentConfigMap) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	if len(c1.ConfigMap) != len(c2.ConfigMap) {
		return false
	}
	for k, f1 := range c1.ConfigMap {
		f2 := c2.ConfigMap[k]
		if !isEqualConfigFile(f1, f2) {
			return false
		}
	}
	return true
}

// isEqualConfigFile compares two configuration files
func isEqualConfigFile(f1, f2 *protobufs.AgentConfigFile) bool {
	if f1 == f2 {
		return true
	}
	if f1 == nil || f2 == nil {
		return false
	}
	return bytes.Equal(f1.Body, f2.Body) && f1.ContentType == f2.ContentType
}
