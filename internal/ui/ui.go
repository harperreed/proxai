package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harperreed/proxai/internal/proxy"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("11")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))
)

type model struct {
	server     *proxy.ProxyServer
	textInput  textinput.Model
	err        error
	quitting   bool
	lastOutput string
}

func initialModel(server *proxy.ProxyServer) model {
	ti := textinput.New()
	ti.Placeholder = "Type a command (r: reset, c: clear, q: quit)"
	ti.Focus()

	return model{
		server:    server,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			m.handleCommand(m.textInput.Value())
			m.textInput.SetValue("")
		}

	case tea.WindowSizeMsg:
		// If we set a width on the input, it can overflow our window,
		// so here we'll set the width to the width of the window
		m.textInput.Width = msg.Width - 4

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) handleCommand(cmd string) {
	switch strings.ToLower(cmd) {
	case "r", "reset":
		m.server.ResetCounters()
		m.lastOutput = "Counters reset."
	case "c", "clear":
		m.lastOutput = ""
	case "q", "quit":
		m.quitting = true
	default:
		m.lastOutput = "Unknown command. Available commands: r/reset, c/clear, q/quit"
	}
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	requestCount, tokensCount, totalCost := m.server.GetStats()
	status := fmt.Sprintf("Tokens: %d | Requests: %d | Cost: $%.4f", tokensCount, requestCount, totalCost)

	s := strings.Builder{}
	s.WriteString(titleStyle.Render("OpenAI Proxy Server"))
	s.WriteString("\n\n")
	s.WriteString(statusStyle.Render(status))
	s.WriteString("\n\n")
	s.WriteString(inputStyle.Render(m.textInput.View()))
	s.WriteString("\n\n")
	s.WriteString(m.lastOutput)

	return s.String()
}

type errMsg error

func StartUI(server *proxy.ProxyServer) error {
	p := tea.NewProgram(initialModel(server))
	_, err := p.Run()
	return err
}
