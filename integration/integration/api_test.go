// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/getlawrence/lawrence-oss/internal/api/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIHealth tests the health endpoint
func TestAPIHealth(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	resp, err := ts.GET("/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
}

// TestAPIAgents tests agent-related endpoints
func TestAPIAgents(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Test GET /api/v1/agents (should be empty initially)
	resp, err := ts.GET("/api/v1/agents")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var agentsResp handlers.GetAgentsResponse
	err = json.NewDecoder(resp.Body).Decode(&agentsResp)
	require.NoError(t, err)

	assert.Equal(t, 0, agentsResp.TotalCount)
	assert.Equal(t, 0, agentsResp.ActiveCount)
	assert.Empty(t, agentsResp.Agents)
}

// TestAPIAgentStats tests agent statistics endpoint
func TestAPIAgentStats(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	resp, err := ts.GET("/api/v1/agents/stats")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats handlers.GetAgentStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&stats)
	require.NoError(t, err)

	assert.Equal(t, 0, stats.TotalAgents)
	assert.Equal(t, 0, stats.OnlineAgents)
	assert.Equal(t, 0, stats.OfflineAgents)
}

// TestAPIGroups tests group-related endpoints
func TestAPIGroups(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Test GET /api/v1/groups (should be empty initially)
	resp, err := ts.GET("/api/v1/groups")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var groupsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&groupsResp)
	require.NoError(t, err)

	groups, ok := groupsResp["groups"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, groups)
}

// TestAPICreateGroup tests group creation
func TestAPICreateGroup(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Create a group
	groupData := map[string]interface{}{
		"name": "Test Group",
		"labels": map[string]string{
			"env": "test",
		},
	}

	body, err := json.Marshal(groupData)
	require.NoError(t, err)

	resp, err := ts.POST("/api/v1/groups", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get the created group ID from response
	var createdGroup map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	require.NoError(t, err)

	groupID, ok := createdGroup["id"].(string)
	require.True(t, ok)
	assert.Equal(t, "Test Group", createdGroup["name"])

	// Verify group was created
	resp, err = ts.GET("/api/v1/groups/" + groupID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var group map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&group)
	require.NoError(t, err)

	assert.Equal(t, groupID, group["id"])
	assert.Equal(t, "Test Group", group["name"])
}

// TestAPIConfigs tests configuration endpoints
func TestAPIConfigs(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Test GET /api/v1/configs (should be empty initially)
	resp, err := ts.GET("/api/v1/configs")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var configsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&configsResp)
	require.NoError(t, err)

	// The API now returns a paginated response with "data" field
	configs, ok := configsResp["data"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, configs)
}

// TestAPITelemetryEndpoints tests telemetry query endpoints
func TestAPITelemetryEndpoints(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Test telemetry overview
	resp, err := ts.GET("/api/v1/telemetry/overview")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test services endpoint
	resp, err = ts.GET("/api/v1/telemetry/services")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestAPITopology tests topology endpoints
func TestAPITopology(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	resp, err := ts.GET("/api/v1/topology")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var topology map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&topology)
	require.NoError(t, err)

	// Should have nodes and edges arrays
	assert.Contains(t, topology, "nodes")
	assert.Contains(t, topology, "edges")
}

// TestAPICORS tests CORS headers
func TestAPICORS(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Make OPTIONS request
	req, err := http.NewRequest("OPTIONS", ts.baseURL+"/api/v1/agents", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check CORS headers
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
}

// TestAPINotFound tests 404 handling
func TestAPINotFound(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	resp, err := ts.GET("/api/v1/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	// API endpoints that don't exist should return 404
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
