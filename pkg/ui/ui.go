package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"druk/pkg/loadtest"
	"druk/pkg/metrics"
)

type Model struct {
	Endpoint  string
	Spinner   spinner.Model
	TextInput textinput.Model
	Metrics   metrics.Metrics
	Err       error
}

func InitialModel(endpoint string) Model {
	return Model{
		Endpoint:  endpoint,
		Spinner:   spinner.New(),
		TextInput: textinput.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		func() tea.Msg {
			metrics, err := loadtest.Run(m.Endpoint, 10*time.Second, 10)
			return loadTestResultMsg{metrics: metrics, err: err}
		},
	)
}

type loadTestResultMsg struct {
	metrics metrics.Metrics
	err     error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	case loadTestResultMsg:
		m.Metrics = msg.metrics
		m.Err = msg.err
		return m, nil
	default:
		return m, nil
	}
	return m, cmd
}

func (m Model) View() string {
	var s string
	if m.Err != nil {
		s += "Error: " + m.Err.Error() + "\n\n"
	}
	s += "Endpoint: " + m.Endpoint + "\n\n"

	latencyPercentiles := m.Metrics.CalculateLatencyPercentiles()

	s += "Metrics:\n"
	s += fmt.Sprintf("Throughput: %.2f requests/second\n", m.Metrics.Throughput)
	s += fmt.Sprintf("Error Rate: %.2f%%\n", m.Metrics.ErrorRate)
	s += fmt.Sprintf("Latency (p90): %s\n", latencyPercentiles.P90)
	s += fmt.Sprintf("Latency (p95): %s\n", latencyPercentiles.P95)
	s += fmt.Sprintf("Latency (p99): %s\n", latencyPercentiles.P99)
	s += fmt.Sprintf("CPU Usage: %.2f%%\n", m.Metrics.CPU)
	s += fmt.Sprintf("Memory Usage: %.2f MB\n", m.Metrics.Memory)
	s += fmt.Sprintf("Network Usage: %.2f Mbps\n", m.Metrics.Network)
	return lipgloss.NewStyle().Render(s)
}
func NewProgram(endpoint string) *tea.Program {
	return tea.NewProgram(InitialModel(endpoint))
}
