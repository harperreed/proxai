package main

import (
	"testing"
)

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		tokens   int
		expected float64
	}{
		{"GPT-3.5-Turbo", "gpt-3.5-turbo", 1000, 0.002},
		{"GPT-4", "gpt-4", 1000, 0.06},
		{"Unknown Model", "unknown-model", 1000, 0.01},
		{"Zero Tokens", "gpt-3.5-turbo", 0, 0},
		{"Large Token Count", "gpt-3.5-turbo", 1000000, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCost(tt.model, tt.tokens)
			if result != tt.expected {
				t.Errorf("calculateCost(%s, %d) = %f, want %f", tt.model, tt.tokens, result, tt.expected)
			}
		})
	}
}
