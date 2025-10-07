// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package query

import (
	"testing"
	"time"
)

func TestParser_ParseSimpleMetricsQuery(t *testing.T) {
	input := `metrics{service="api"} [5m]`
	parser := NewParser(input)
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	telemetryQuery, ok := query.(*TelemetryQuery)
	if !ok {
		t.Fatalf("Expected TelemetryQuery, got %T", query)
	}

	if telemetryQuery.Type != TelemetryTypeMetrics {
		t.Errorf("Expected type 'metrics', got %s", telemetryQuery.Type)
	}

	if telemetryQuery.Duration != 5*time.Minute {
		t.Errorf("Expected duration 5m, got %v", telemetryQuery.Duration)
	}

	serviceSelector, exists := telemetryQuery.Selectors["service"]
	if !exists {
		t.Fatal("Expected 'service' selector")
	}

	if serviceSelector.Operator != SelectorOpEqual {
		t.Errorf("Expected operator '=', got %s", serviceSelector.Operator)
	}

	if serviceSelector.Value != "api" {
		t.Errorf("Expected value 'api', got %s", serviceSelector.Value)
	}
}

func TestParser_ParseLogsQuery(t *testing.T) {
	input := `logs{severity="error", body=~".*timeout.*"} [1h]`
	parser := NewParser(input)
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	telemetryQuery, ok := query.(*TelemetryQuery)
	if !ok {
		t.Fatalf("Expected TelemetryQuery, got %T", query)
	}

	if telemetryQuery.Type != TelemetryTypeLogs {
		t.Errorf("Expected type 'logs', got %s", telemetryQuery.Type)
	}

	severitySelector, exists := telemetryQuery.Selectors["severity"]
	if !exists {
		t.Fatal("Expected 'severity' selector")
	}

	if severitySelector.Value != "error" {
		t.Errorf("Expected severity value 'error', got %s", severitySelector.Value)
	}

	bodySelector, exists := telemetryQuery.Selectors["body"]
	if !exists {
		t.Fatal("Expected 'body' selector")
	}

	if bodySelector.Operator != SelectorOpRegex {
		t.Errorf("Expected regex operator, got %s", bodySelector.Operator)
	}
}

func TestParser_ParseTracesQuery(t *testing.T) {
	input := `traces{service="api", trace_id="abc123"} [30m]`
	parser := NewParser(input)
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	telemetryQuery, ok := query.(*TelemetryQuery)
	if !ok {
		t.Fatalf("Expected TelemetryQuery, got %T", query)
	}

	if telemetryQuery.Type != TelemetryTypeTraces {
		t.Errorf("Expected type 'traces', got %s", telemetryQuery.Type)
	}

	if len(telemetryQuery.Selectors) != 2 {
		t.Errorf("Expected 2 selectors, got %d", len(telemetryQuery.Selectors))
	}
}

func TestParser_ParseFunctionCall(t *testing.T) {
	input := `sum(metrics{metric="cpu_usage"} [5m])`
	parser := NewParser(input)
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	functionCall, ok := query.(*FunctionCall)
	if !ok {
		t.Fatalf("Expected FunctionCall, got %T", query)
	}

	if functionCall.Name != "sum" {
		t.Errorf("Expected function name 'sum', got %s", functionCall.Name)
	}

	if len(functionCall.Args) != 1 {
		t.Fatalf("Expected 1 argument, got %d", len(functionCall.Args))
	}

	arg, ok := functionCall.Args[0].(*TelemetryQuery)
	if !ok {
		t.Fatalf("Expected argument to be TelemetryQuery, got %T", functionCall.Args[0])
	}

	if arg.Type != TelemetryTypeMetrics {
		t.Errorf("Expected argument type 'metrics', got %s", arg.Type)
	}
}

func TestParser_ParseAggregation(t *testing.T) {
	input := `avg(metrics{metric="response_time"} [1h]) by (service)`
	parser := NewParser(input)
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	aggregation, ok := query.(*Aggregation)
	if !ok {
		t.Fatalf("Expected Aggregation, got %T", query)
	}

	if aggregation.Function != "avg" {
		t.Errorf("Expected function 'avg', got %s", aggregation.Function)
	}

	if len(aggregation.By) != 1 {
		t.Fatalf("Expected 1 grouping label, got %d", len(aggregation.By))
	}

	if aggregation.By[0] != "service" {
		t.Errorf("Expected grouping by 'service', got %s", aggregation.By[0])
	}
}

func TestParser_ParseBinaryOp(t *testing.T) {
	input := `metrics{service="api"} [5m] + metrics{service="web"} [5m]`
	parser := NewParser(input)
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	binaryOp, ok := query.(*BinaryOp)
	if !ok {
		t.Fatalf("Expected BinaryOp, got %T", query)
	}

	if binaryOp.Operator != BinaryOpAdd {
		t.Errorf("Expected operator '+', got %s", binaryOp.Operator)
	}

	leftQuery, ok := binaryOp.Left.(*TelemetryQuery)
	if !ok {
		t.Fatalf("Expected left operand to be TelemetryQuery, got %T", binaryOp.Left)
	}

	rightQuery, ok := binaryOp.Right.(*TelemetryQuery)
	if !ok {
		t.Fatalf("Expected right operand to be TelemetryQuery, got %T", binaryOp.Right)
	}

	if leftQuery.Selectors["service"].Value != "api" {
		t.Errorf("Expected left query service 'api', got %s", leftQuery.Selectors["service"].Value)
	}

	if rightQuery.Selectors["service"].Value != "web" {
		t.Errorf("Expected right query service 'web', got %s", rightQuery.Selectors["service"].Value)
	}
}

func TestParser_InvalidSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing braces", "metrics [5m]"},
		{"invalid operator", "metrics{service#\"api\"} [5m]"},
		{"unterminated string", "metrics{service=\"api} [5m]"},
		{"unknown type", "unknown{} [5m]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			_, err := parser.Parse()
			if err == nil {
				t.Errorf("Expected error for invalid syntax: %s", tt.input)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"5s", 5 * time.Second},
		{"10m", 10 * time.Minute},
		{"2h", 2 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			duration, err := parseDuration(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse duration: %v", err)
			}
			if duration != tt.expected {
				t.Errorf("Expected duration %v, got %v", tt.expected, duration)
			}
		})
	}
}

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"valid metrics query", `metrics{service="api"} [5m]`, false},
		{"valid logs query", `logs{severity="error"} [1h]`, false},
		{"valid function call", `sum(metrics{} [5m])`, false},
		{"invalid syntax", `metrics{service= [5m]`, true},
		{"empty query", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
