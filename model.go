package main

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    // "github.com/charmbracelet/lipgloss"
)

type model struct {
    viewport viewport.Model
    content  string
    ready    bool
    service  Service  // Add service to the model
}

func initialModel() model {
    return model{
        viewport: viewport.New(80, 20),
        content:  "",
        ready:    false,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyCtrlC:
            return m, tea.Quit
        case tea.KeyRunes:
            switch string(msg.Runes) {
            case "r", "R":
                requestCount = 0
                tokensCount = 0
                totalCost = 0.0
                m.content += "\nCounters reset."
            case "c", "C":
                cmd = tea.Sequence(
                    tea.Printf("\033[2J"),
                    tea.Printf("\033[1;1H"),
                )
                m.content = ""
            }
        }

    case tea.WindowSizeMsg:
        m.viewport.Width = msg.Width
        // m.viewport.Height = msg.Height - lipgloss.Height(statusBar())

    case string:
        m.content += msg
    }

    m.viewport.SetContent(m.content)
    m.viewport, _ = m.viewport.Update(msg)

    // Adjust the viewport's YOffset to scroll to the bottom
    lines := strings.Split(m.content, "\n")
    if len(lines) > m.viewport.Height {
        m.viewport.YOffset = len(lines) - m.viewport.Height
    }

    return m, cmd
}

func (m model) View() string {
    return fmt.Sprintf("%s\n%s", m.viewport.View(), m.service.GetStatus())
}
