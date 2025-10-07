// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/getlawrence/lawrence-oss/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

func setupAgentHandlersTest() (*AgentHandlers, *testutils.MockAgentService) {
	mockService := testutils.NewMockAgentService()
	logger := zap.NewNop()
	handlers := NewAgentHandlers(mockService, logger)
	return handlers, mockService
}

func TestHandleGetAgents(t *testing.T) {
	handlers, mockService := setupAgentHandlersTest()

	// Create test agents
	agent1 := testutils.MakeTestAgentWithStatus(uuid.New(), services.AgentStatusOnline)
	agent2 := testutils.MakeTestAgentWithStatus(uuid.New(), services.AgentStatusOffline)

	_ = mockService.CreateAgent(context.TODO(), agent1)
	_ = mockService.CreateAgent(context.TODO(), agent2)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/agents", nil)

	// Execute handler
	handlers.HandleGetAgents(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response GetAgentsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, 1, response.ActiveCount)
	assert.Equal(t, 1, response.InactiveCount)
	assert.Len(t, response.Agents, 2)
}

func TestHandleGetAgents_ServiceError(t *testing.T) {
	handlers, mockService := setupAgentHandlersTest()

	// Set error flag
	mockService.ListAgentsErr = fmt.Errorf("database error")

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/agents", nil)

	// Execute handler
	handlers.HandleGetAgents(c)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to fetch agents")
}

func TestHandleGetAgent(t *testing.T) {
	handlers, mockService := setupAgentHandlersTest()

	// Create test agent
	agentID := uuid.New()
	agent := testutils.MakeTestAgent(agentID)
	_ = mockService.CreateAgent(context.TODO(), agent)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/agents/%s", agentID), nil)
	c.Params = gin.Params{{Key: "id", Value: agentID.String()}}

	// Execute handler
	handlers.HandleGetAgent(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response services.Agent
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, agentID, response.ID)
	assert.Equal(t, agent.Name, response.Name)
}

func TestHandleGetAgent_NotFound(t *testing.T) {
	handlers, _ := setupAgentHandlersTest()

	// Use non-existent agent ID
	nonExistentID := uuid.New()

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/agents/%s", nonExistentID), nil)
	c.Params = gin.Params{{Key: "id", Value: nonExistentID.String()}}

	// Execute handler
	handlers.HandleGetAgent(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Agent not found")
}

func TestHandleGetAgent_InvalidID(t *testing.T) {
	handlers, _ := setupAgentHandlersTest()

	// Create test request with invalid UUID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/agents/invalid-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	// Execute handler
	handlers.HandleGetAgent(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid agent ID format")
}

func TestHandleGetAgent_MissingID(t *testing.T) {
	handlers, _ := setupAgentHandlersTest()

	// Create test request without ID param
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/agents/", nil)

	// Execute handler
	handlers.HandleGetAgent(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Agent ID is required")
}

func TestHandleGetAgentStats(t *testing.T) {
	handlers, mockService := setupAgentHandlersTest()

	// Create test agents with different statuses
	agent1 := testutils.MakeTestAgentWithStatus(uuid.New(), services.AgentStatusOnline)
	agent2 := testutils.MakeTestAgentWithStatus(uuid.New(), services.AgentStatusOnline)
	agent3 := testutils.MakeTestAgentWithStatus(uuid.New(), services.AgentStatusOffline)
	agent4 := testutils.MakeTestAgentWithStatus(uuid.New(), services.AgentStatusError)

	_ = mockService.CreateAgent(context.TODO(), agent1)
	_ = mockService.CreateAgent(context.TODO(), agent2)
	_ = mockService.CreateAgent(context.TODO(), agent3)
	_ = mockService.CreateAgent(context.TODO(), agent4)

	// Create test groups
	group1 := testutils.MakeTestGroup("group-1")
	group2 := testutils.MakeTestGroup("group-2")
	_ = mockService.CreateGroup(context.TODO(), group1)
	_ = mockService.CreateGroup(context.TODO(), group2)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/agents/stats", nil)

	// Execute handler
	handlers.HandleGetAgentStats(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response GetAgentStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 4, response.TotalAgents)
	assert.Equal(t, 2, response.OnlineAgents)
	assert.Equal(t, 1, response.OfflineAgents)
	assert.Equal(t, 1, response.ErrorAgents)
	assert.Equal(t, 2, response.GroupsCount)
}

func TestHandleGetAgentStats_ServiceError(t *testing.T) {
	handlers, mockService := setupAgentHandlersTest()

	// Set error flag
	mockService.ListAgentsErr = fmt.Errorf("database error")

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/agents/stats", nil)

	// Execute handler
	handlers.HandleGetAgentStats(c)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to fetch agent statistics")
}

func TestHandleGetAgentStats_GroupsError(t *testing.T) {
	handlers, mockService := setupAgentHandlersTest()

	// Create test agent
	agent := testutils.MakeTestAgent(uuid.New())
	_ = mockService.CreateAgent(context.TODO(), agent)

	// Set error flag for groups only
	mockService.ListGroupsErr = fmt.Errorf("groups error")

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/agents/stats", nil)

	// Execute handler
	handlers.HandleGetAgentStats(c)

	// Assert response - should succeed with groups count = 0
	assert.Equal(t, http.StatusOK, w.Code)

	var response GetAgentStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.TotalAgents)
	assert.Equal(t, 0, response.GroupsCount)
}

func TestHandleUpdateAgentGroup(t *testing.T) {
	handlers, _ := setupAgentHandlersTest()

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PATCH", "/api/v1/agents/"+uuid.New().String()+"/group", nil)

	// Execute handler
	handlers.HandleUpdateAgentGroup(c)

	// Assert response - not implemented
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}
