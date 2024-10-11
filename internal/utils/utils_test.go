package utils

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
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
		{"Negative Tokens", "gpt-3.5-turbo", -1000, 0},
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

func TestLipglossStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
		check func(lipgloss.Style) bool
	}{
		{"InfoStyle", InfoStyle, func(s lipgloss.Style) bool { return s.GetForeground().String() == "14" }},
		{"SuccessStyle", SuccessStyle, func(s lipgloss.Style) bool { return s.GetForeground().String() == "10" }},
		{"ErrorStyle", ErrorStyle, func(s lipgloss.Style) bool { return s.GetForeground().String() == "9" }},
		{"BoldStyle", BoldStyle, func(s lipgloss.Style) bool { return s.GetBold() }},
		{"LipglossStyle", LipglossStyle, func(s lipgloss.Style) bool { return s.GetForeground().String() == "241" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.style) {
				t.Errorf("%s is not correctly initialized", tt.name)
			}
		})
	}
}
