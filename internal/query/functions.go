// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package query

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Function represents a built-in query function
type Function struct {
	Name        string
	Description string
	Apply       func([]QueryResult) (interface{}, error)
}

var functions = map[string]*Function{
	"sum":                sumFunction(),
	"avg":                avgFunction(),
	"min":                minFunction(),
	"max":                maxFunction(),
	"count":              countFunction(),
	"rate":               rateFunction(),
	"increase":           increaseFunction(),
	"histogram_quantile": histogramQuantileFunction(),
}

// GetFunction returns a function by name
func GetFunction(name string) (*Function, bool) {
	fn, ok := functions[name]
	return fn, ok
}

// ListFunctions returns all available functions
func ListFunctions() []*Function {
	fns := make([]*Function, 0, len(functions))
	for _, fn := range functions {
		fns = append(fns, fn)
	}
	return fns
}

// sumFunction returns the sum aggregation function
func sumFunction() *Function {
	return &Function{
		Name:        "sum",
		Description: "Calculates the sum of values",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return []QueryResult{}, nil
			}

			sum := 0.0
			labels := make(map[string]string)
			var timestamp time.Time

			for _, r := range results {
				// Convert value to float64
				val, err := toFloat64(r.Value)
				if err != nil {
					continue
				}
				sum += val

				// Use labels from first result
				if len(labels) == 0 {
					labels = r.Labels
					timestamp = r.Timestamp
				}
			}

			return []QueryResult{{
				Type:      results[0].Type,
				Timestamp: timestamp,
				Labels:    labels,
				Value:     sum,
				Data: map[string]interface{}{
					"function": "sum",
					"count":    len(results),
				},
			}}, nil
		},
	}
}

// avgFunction returns the average aggregation function
func avgFunction() *Function {
	return &Function{
		Name:        "avg",
		Description: "Calculates the average of values",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return []QueryResult{}, nil
			}

			sum := 0.0
			count := 0
			labels := make(map[string]string)
			var timestamp time.Time

			for _, r := range results {
				val, err := toFloat64(r.Value)
				if err != nil {
					continue
				}
				sum += val
				count++

				if len(labels) == 0 {
					labels = r.Labels
					timestamp = r.Timestamp
				}
			}

			if count == 0 {
				return []QueryResult{}, nil
			}

			return []QueryResult{{
				Type:      results[0].Type,
				Timestamp: timestamp,
				Labels:    labels,
				Value:     sum / float64(count),
				Data: map[string]interface{}{
					"function": "avg",
					"count":    count,
				},
			}}, nil
		},
	}
}

// minFunction returns the minimum aggregation function
func minFunction() *Function {
	return &Function{
		Name:        "min",
		Description: "Returns the minimum value",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return []QueryResult{}, nil
			}

			min := math.Inf(1)
			var minResult QueryResult

			for _, r := range results {
				val, err := toFloat64(r.Value)
				if err != nil {
					continue
				}
				if val < min {
					min = val
					minResult = r
				}
			}

			if math.IsInf(min, 1) {
				return []QueryResult{}, nil
			}

			minResult.Value = min
			minResult.Data = map[string]interface{}{
				"function": "min",
			}

			return []QueryResult{minResult}, nil
		},
	}
}

// maxFunction returns the maximum aggregation function
func maxFunction() *Function {
	return &Function{
		Name:        "max",
		Description: "Returns the maximum value",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return []QueryResult{}, nil
			}

			max := math.Inf(-1)
			var maxResult QueryResult

			for _, r := range results {
				val, err := toFloat64(r.Value)
				if err != nil {
					continue
				}
				if val > max {
					max = val
					maxResult = r
				}
			}

			if math.IsInf(max, -1) {
				return []QueryResult{}, nil
			}

			maxResult.Value = max
			maxResult.Data = map[string]interface{}{
				"function": "max",
			}

			return []QueryResult{maxResult}, nil
		},
	}
}

// countFunction returns the count aggregation function
func countFunction() *Function {
	return &Function{
		Name:        "count",
		Description: "Counts the number of results",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return []QueryResult{}, nil
			}

			labels := make(map[string]string)
			var timestamp time.Time

			if len(results) > 0 {
				labels = results[0].Labels
				timestamp = results[0].Timestamp
			}

			return []QueryResult{{
				Type:      results[0].Type,
				Timestamp: timestamp,
				Labels:    labels,
				Value:     float64(len(results)),
				Data: map[string]interface{}{
					"function": "count",
				},
			}}, nil
		},
	}
}

// rateFunction calculates the per-second rate of increase
func rateFunction() *Function {
	return &Function{
		Name:        "rate",
		Description: "Calculates the per-second rate of increase (only works with metrics)",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return nil, fmt.Errorf("rate() requires at least one result")
			}

			// Check if this is metrics data
			if results[0].Type != "metrics" {
				return nil, fmt.Errorf("rate() function can only be used with metrics, not %s", results[0].Type)
			}

			if len(results) < 2 {
				return nil, fmt.Errorf("rate() requires at least 2 data points, got %d", len(results))
			}

			// Sort by timestamp
			sort.Slice(results, func(i, j int) bool {
				return results[i].Timestamp.Before(results[j].Timestamp)
			})

			first := results[0]
			last := results[len(results)-1]

			firstVal, err := toFloat64(first.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert first value to number: %v", err)
			}

			lastVal, err := toFloat64(last.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert last value to number: %v", err)
			}

			timeDiff := last.Timestamp.Sub(first.Timestamp).Seconds()
			if timeDiff == 0 {
				return nil, fmt.Errorf("rate() requires time difference between data points")
			}

			rate := (lastVal - firstVal) / timeDiff

			return []QueryResult{{
				Type:      first.Type,
				Timestamp: last.Timestamp,
				Labels:    first.Labels,
				Value:     rate,
				Data: map[string]interface{}{
					"function":  "rate",
					"time_span": timeDiff,
				},
			}}, nil
		},
	}
}

// increaseFunction calculates the total increase over time range
func increaseFunction() *Function {
	return &Function{
		Name:        "increase",
		Description: "Calculates the total increase over the time range (only works with metrics)",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return nil, fmt.Errorf("increase() requires at least one result")
			}

			// Check if this is metrics data
			if results[0].Type != "metrics" {
				return nil, fmt.Errorf("increase() function can only be used with metrics, not %s", results[0].Type)
			}

			if len(results) < 2 {
				return nil, fmt.Errorf("increase() requires at least 2 data points, got %d", len(results))
			}

			// Sort by timestamp
			sort.Slice(results, func(i, j int) bool {
				return results[i].Timestamp.Before(results[j].Timestamp)
			})

			first := results[0]
			last := results[len(results)-1]

			firstVal, err := toFloat64(first.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert first value to number: %v", err)
			}

			lastVal, err := toFloat64(last.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert last value to number: %v", err)
			}

			increase := lastVal - firstVal

			return []QueryResult{{
				Type:      first.Type,
				Timestamp: last.Timestamp,
				Labels:    first.Labels,
				Value:     increase,
				Data: map[string]interface{}{
					"function": "increase",
				},
			}}, nil
		},
	}
}

// histogramQuantileFunction calculates histogram quantiles
func histogramQuantileFunction() *Function {
	return &Function{
		Name:        "histogram_quantile",
		Description: "Calculates histogram quantile (e.g., p95, p99)",
		Apply: func(results []QueryResult) (interface{}, error) {
			if len(results) == 0 {
				return []QueryResult{}, nil
			}

			// Extract values and sort
			values := make([]float64, 0, len(results))
			for _, r := range results {
				val, err := toFloat64(r.Value)
				if err != nil {
					continue
				}
				values = append(values, val)
			}

			if len(values) == 0 {
				return []QueryResult{}, nil
			}

			sort.Float64s(values)

			// Calculate common quantiles
			quantiles := map[string]float64{
				"p50": calculateQuantile(values, 0.5),
				"p90": calculateQuantile(values, 0.9),
				"p95": calculateQuantile(values, 0.95),
				"p99": calculateQuantile(values, 0.99),
			}

			// Return results for each quantile
			quantileResults := make([]QueryResult, 0, len(quantiles))
			for label, value := range quantiles {
				labels := make(map[string]string)
				if len(results) > 0 && results[0].Labels != nil {
					for k, v := range results[0].Labels {
						labels[k] = v
					}
				}
				labels["quantile"] = label

				quantileResults = append(quantileResults, QueryResult{
					Type:      results[0].Type,
					Timestamp: results[0].Timestamp,
					Labels:    labels,
					Value:     value,
					Data: map[string]interface{}{
						"function": "histogram_quantile",
					},
				})
			}

			return quantileResults, nil
		},
	}
}

// calculateQuantile calculates the quantile value from sorted values
func calculateQuantile(sortedValues []float64, quantile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := quantile * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

// toFloat64 converts various types to float64
func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case string:
		// Try to parse as float
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}

// GetFunctionSuggestions returns function suggestions for auto-completion
func GetFunctionSuggestions() []string {
	suggestions := make([]string, 0, len(functions))
	for name := range functions {
		suggestions = append(suggestions, name+"(")
	}
	sort.Strings(suggestions)
	return suggestions
}
