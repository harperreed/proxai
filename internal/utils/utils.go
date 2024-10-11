package utils

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	InfoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	SuccessStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	ErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	BoldStyle     = lipgloss.NewStyle().Bold(true)
	LipglossStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func calculateCost(model string, tokens int) float64 {
	// These are example prices, you should update them with the actual pricing
	prices := map[string]float64{
		"gpt-3.5-turbo": 0.002 / 1000,
		"gpt-4":         0.06 / 1000,
		"default":       0.01 / 1000,
	}

	price, ok := prices[model]
	if !ok {
		price = prices["default"]
	}

	return float64(tokens) * price
}
