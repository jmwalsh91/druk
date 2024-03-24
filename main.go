package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	endpoint  string
	spinner   spinner.Model
	textInput textinput.Model
	metrics   metrics
	err       error
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runLoadTest,
	)
}

type loadTestResultMsg struct {
	metrics metrics
	err     error
}

func (m model) runLoadTest() tea.Msg {
	time.Sleep(2 * time.Second)

	return loadTestResultMsg{
		metrics: metrics{
			throughput: 100.0,
			errorRate:  1.5,
			latency: struct {
				p90 time.Duration
				p95 time.Duration
				p99 time.Duration
			}{
				p90: 100 * time.Millisecond,
				p95: 150 * time.Millisecond,
				p99: 200 * time.Millisecond,
			},
			cpu:     20.0,
			memory:  100.0,
			network: 10.0,
		},
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case loadTestResultMsg:
		m.metrics = msg.metrics
		m.err = msg.err
		return m, nil

	default:
		return m, nil
	}

	return m, cmd
}

func (m model) View() string {
	var s string

	if m.err != nil {
		s += "Error: " + m.err.Error() + "\n\n"
	}

	s += "Endpoint: " + m.endpoint + "\n\n"

	s += "Metrics:\n"
	s += fmt.Sprintf("Throughput: %.2f requests/second\n", m.metrics.throughput)
	s += fmt.Sprintf("Error Rate: %.2f%%\n", m.metrics.errorRate)
	s += fmt.Sprintf("Latency (p90): %s\n", m.metrics.latency.p90)
	s += fmt.Sprintf("Latency (p95): %s\n", m.metrics.latency.p95)
	s += fmt.Sprintf("Latency (p99): %s\n", m.metrics.latency.p99)
	s += fmt.Sprintf("CPU Usage: %.2f%%\n", m.metrics.cpu)
	s += fmt.Sprintf("Memory Usage: %.2f MB\n", m.metrics.memory)
	s += fmt.Sprintf("Network Usage: %.2f Mbps\n", m.metrics.network)

	return lipgloss.NewStyle().Render(s)
}

type metrics struct {
	throughput float64
	errorRate  float64
	latency    struct {
		p90 time.Duration
		p95 time.Duration
		p99 time.Duration
	}
	cpu     float64
	memory  float64
	network float64
}

func initialModel(endpoint string) model {
	return model{
		endpoint:  endpoint,
		spinner:   spinner.New(),
		textInput: textinput.New(),
	}
}

func main() {
	endpoint := flag.String("endpoint", "", "API endpoint to test")
	flag.Parse()

	if *endpoint == "" {
		fmt.Println("Please provide an endpoint using the -endpoint flag")
		return
	}

	p := tea.NewProgram(initialModel(*endpoint))

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
