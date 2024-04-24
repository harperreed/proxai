package main

import "github.com/charmbracelet/lipgloss"

var (
    infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
    successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
    errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
    boldStyle    = lipgloss.NewStyle().Bold(true)
    statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("11")).Padding(0, 1)
)
