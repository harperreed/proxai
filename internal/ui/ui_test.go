package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/harperreed/proxai/internal/proxy"
)

// MockProxyServer is a mock implementation of proxy.ProxyServer
type MockProxyServer struct {
	RequestCount int64
	TokensCount  int64
	TotalCost    float64
}

func (m *MockProxyServer) ResetCounters() {
	m.RequestCount = 0
	m.TokensCount = 0
	m.TotalCost = 0
}

func (m *MockProxyServer) GetStats() (int64, int64, float64) {
	return m.RequestCount, m.TokensCount, m.TotalCost
}

func TestInitialModel(t *testing.T) {
	server := &MockProxyServer{}
	m := initialModel(server)

	if m.server != server {
		t.Errorf("Expected server to be %v, got %v", server, m.server)
	}

	if m.textInput.Placeholder != "Type a command (r: reset, c: clear, q: quit)" {
		t.Errorf("Unexpected placeholder: %s", m.textInput.Placeholder)
	}

	if !m.textInput.Focused() {
		t.Error("Expected textInput to be focused")
	}
}

func TestModelUpdate(t *testing.T) {
	server := &MockProxyServer{}
	m := initialModel(server)

	tests := []struct {
		name     string
		msg      tea.Msg
		expected model
	}{
		{
			name: "Quit on Ctrl+C",
			msg:  tea.KeyMsg{Type: tea.KeyCtrlC},
			expected: model{
				server:   server,
				quitting: true,
			},
		},
		{
			name: "Quit on Esc",
			msg:  tea.KeyMsg{Type: tea.KeyEsc},
			expected: model{
				server:   server,
				quitting: true,
			},
		},
		{
			name: "Handle Enter key",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
			expected: model{
				server:     server,
				lastOutput: "Unknown command. Available commands: r/reset, c/clear, q/quit",
			},
		},
		{
			name: "Handle window size change",
			msg:  tea.WindowSizeMsg{Width: 100, Height: 50},
			expected: model{
				server: server,
				textInput: func() textinput.Model {
					ti := textinput.New()
					ti.Width = 96 // 100 - 4
					return ti
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := m.Update(tt.msg)
			updatedModel := newModel.(model)

			if updatedModel.quitting != tt.expected.quitting {
				t.Errorf("Expected quitting to be %v, got %v", tt.expected.quitting, updatedModel.quitting)
			}

			if updatedModel.lastOutput != tt.expected.lastOutput {
				t.Errorf("Expected lastOutput to be %q, got %q", tt.expected.lastOutput, updatedModel.lastOutput)
			}

			if tt.msg == tea.WindowSizeMsg {
				if updatedModel.textInput.Width != tt.expected.textInput.Width {
					t.Errorf("Expected textInput width to be %d, got %d", tt.expected.textInput.Width, updatedModel.textInput.Width)
				}
			}
		})
	}
}

func TestHandleCommand(t *testing.T) {
	server := &MockProxyServer{RequestCount: 10, TokensCount: 100, TotalCost: 1.5}
	m := initialModel(server)

	tests := []struct {
		name           string
		command        string
		expectedOutput string
		expectedReset  bool
	}{
		{"Reset command", "r", "Counters reset.", true},
		{"Reset command (full)", "reset", "Counters reset.", true},
		{"Clear command", "c", "", false},
		{"Clear command (full)", "clear", "", false},
		{"Quit command", "q", "", false},
		{"Quit command (full)", "quit", "", false},
		{"Unknown command", "unknown", "Unknown command. Available commands: r/reset, c/clear, q/quit", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.handleCommand(tt.command)

			if m.lastOutput != tt.expectedOutput {
				t.Errorf("Expected output %q, got %q", tt.expectedOutput, m.lastOutput)
			}

			if tt.expectedReset {
				if server.RequestCount != 0 || server.TokensCount != 0 || server.TotalCost != 0 {
					t.Error("Expected server counters to be reset")
				}
			}

			if tt.command == "q" || tt.command == "quit" {
				if !m.quitting {
					t.Error("Expected quitting to be true")
				}
			}
		})
	}
}

func TestView(t *testing.T) {
	server := &MockProxyServer{RequestCount: 10, TokensCount: 100, TotalCost: 1.5}
	m := initialModel(server)

	view := m.View()

	expectedSubstrings := []string{
		"OpenAI Proxy Server",
		"Tokens: 100 | Requests: 10 | Cost: $1.5000",
		"Type a command (r: reset, c: clear, q: quit)",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(view, substr) {
			t.Errorf("Expected view to contain %q", substr)
		}
	}

	m.quitting = true
	quitView := m.View()
	if quitView != "Goodbye!\n" {
		t.Errorf("Expected quit view to be 'Goodbye!\\n', got %q", quitView)
	}
}

func TestStartUI(t *testing.T) {
	server := &MockProxyServer{}

	// Mock tea.NewProgram to avoid actually starting the UI
	originalNewProgram := tea.NewProgram
	tea.NewProgram = func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		return &tea.Program{}
	}
	defer func() { tea.NewProgram = originalNewProgram }()

	err := StartUI(server)
	if err != nil {
		t.Errorf("StartUI returned unexpected error: %v", err)
	}
}
