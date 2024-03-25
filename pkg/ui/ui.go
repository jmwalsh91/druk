package ui

import (
	"fmt"
	"image"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"druk/pkg/loadtest"
	"druk/pkg/metrics"
)

var (
	columnWidth = 30
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderBackground(lipgloss.Color("63")).Padding(1, 1)
	statusCodeStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBackground(lipgloss.Color("63")).
			BorderBottom(true).
			Width(columnWidth).
			Align(lipgloss.Center)
	errorStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			BorderBottom(true).
			Width(columnWidth).
			Align(lipgloss.Center)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Italic(true).
			Foreground(lipgloss.Color("#FFC0CB")).
			PaddingTop(1).
			PaddingBottom(1).
			PaddingLeft(2).
			PaddingRight(2).
			Align(lipgloss.Center)

	drukArt = `
    ▓█████▄  ██▀███   █    ██  ██ ▄█▀
    ▒██▀ ██▌▓██ ▒ ██▒ ██  ▓██▒ ██▄█▒ 
    ░██   █▌▓██ ░▄█ ▒▓██  ▒██░▓███▄░ 
    ░▓█▄   ▌▒██▀▀█▄  ▓▓█  ░██░▓██ █▄ 
    ░▒████▓ ░██▓ ▒██▒▒▒█████▓ ▒██▒ █▄
     ▒▒▓  ▒ ░ ▒▓ ░▒▓░░▒▓▒ ▒ ▒ ▒ ▒▒ ▓▒
     ░ ▒  ▒   ░▒ ░ ▒░░░▒░ ░ ░ ░ ░▒ ▒░
     ░ ░  ░   ░░   ░  ░░░ ░ ░ ░ ░░ ░ 
       ░       ░        ░     ░  ░   
     ░                               
    `
)

func formatDuration(d time.Duration) string {
	if d == 0 {
		log.Println("formatDuration: Duration is zero")
		return "0s"
	}
	log.Printf("formatDuration: Formatting duration: %v", d)
	return d.String()
}

type Model struct {
	Endpoint        string
	Duration        time.Duration
	Concurrency     int
	Status          string
	Progress        float64
	ProgressBar     progress.Model
	Metrics         metrics.Metrics
	LatencyGraph    *LineChart
	ThroughputGraph *LineChart
	Err             error
	Quitting        bool
}

type loadTestResultMsg struct {
	metrics metrics.Metrics
	err     error
}

func InitialModel(endpoint string, duration time.Duration, concurrency int) Model {
	return Model{
		Endpoint:    endpoint,
		Duration:    duration,
		Concurrency: concurrency,
		Status:      "Ready",
		ProgressBar: progress.New(progress.WithGradient("00FFFF", "FF00FF")),
		LatencyGraph: &LineChart{
			Title: "Latency",
			Data:  []float64{},
		},
		ThroughputGraph: &LineChart{
			Title: "Throughput",
			Data:  []float64{},
		},
		Metrics: metrics.Metrics{
			StatusCodes: make(map[int]int),
			Errors:      make(map[string]int),
		},
	}
}
func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.Quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			if m.Status == "Ready" {
				m.Status = "Running"
				cmd = m.runLoadTest()
			}
		}

	case tea.WindowSizeMsg:
		m.ProgressBar.Width = msg.Width - 4
		m.LatencyGraph.Width = msg.Width/2 - 4
		m.ThroughputGraph.Width = msg.Width/2 - 4

	case progress.FrameMsg:
		progressModel, cmd := m.ProgressBar.Update(msg)
		m.ProgressBar = progressModel.(progress.Model)
		return m, cmd

	case metrics.Metrics:
		log.Printf("Received metrics: %+v", msg)
		m.Metrics = msg
		m.Metrics.CalculateStatistics()
		m.updateLineCharts()

	case loadTestResultMsg:
		m.Metrics = msg.metrics
		m.Err = msg.err
		m.Status = "Completed"
		return m, tea.Quit
	}

	return m, cmd
}
func (m Model) View() string {
	var sections []string

	// Progress Bar
	durationProgress, err := safeDurationConversion(m.Progress)
	if err != nil {
		log.Printf("Error converting progress duration: %v", err)
	}

	var progress string
	if m.Duration == 0 {
		progress = "Duration is not set"
	} else {
		progress = lipgloss.NewStyle().Render(fmt.Sprintf("%s / %s", formatDuration(durationProgress), formatDuration(m.Duration)))
	}
	sections = append(sections, progress)

	// Stats for Last Second
	statsSection, err := m.renderStatsSection()
	if err != nil {
		log.Printf("Error rendering stats section: %v", err)
	} else {
		sections = append(sections, columnStyle.Render(statsSection))
	}

	// Status Code Distribution
	statusCodeSection, err := m.renderStatusCodeSection()
	if err != nil {
		log.Printf("Error rendering status code section: %v", err)
	} else {
		sections = append(sections, columnStyle.Render(statusCodeSection))
	}

	// Error Distribution
	errorSection, err := m.renderErrorSection()
	if err != nil {
		log.Printf("Error rendering error section: %v", err)
	} else {
		sections = append(sections, columnStyle.Render(errorSection))
	}

	// Line Charts
	lineCharts, err := m.renderLineCharts()
	if err != nil {
		log.Printf("Error rendering line charts: %v", err)
	} else {
		sections = append(sections, lineCharts)
	}

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func safeDurationConversion(f float64) (time.Duration, error) {
	if f < 0 {
		return 0, fmt.Errorf("invalid negative duration: %f", f)
	}
	return time.Duration(f * float64(time.Second)), nil
}
func (m *Model) renderStatsSection() (string, error) {
	if m.Metrics.LatencyData == nil || m.Metrics.LatencyP99 == 0 {
		return "", fmt.Errorf("invalid metrics data")
	}
	if m.Metrics.RequestsPerSecond == 0 {
		return "", fmt.Errorf("invalid metrics data")
	}
	statsSection := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Render("Stats for Last Second"),
		lipgloss.NewStyle().Render(fmt.Sprintf(
			"Requests: %.0f\nSlowest: %.4f secs\nFastest: %.4f secs\nAverage: %.4f secs\nData: %.2f MiB\nNumber of open files: %d / %d",
			m.Metrics.RequestsPerSecond,
			float64(m.Metrics.LatencyP99)/float64(time.Second),
			m.Metrics.LatencyData[0]/float64(time.Millisecond)*float64(time.Second),
			float64(m.Metrics.AvgLatency)/float64(time.Second),
			m.Metrics.DataTransferred,
			m.Metrics.OpenFiles,
			m.Metrics.MaxOpenFiles,
		)),
	)

	return statsSection, nil
}

func (m *Model) renderStatusCodeSection() (string, error) {
	if m.Metrics.StatusCodes == nil {
		return "", fmt.Errorf("status codes map is nil")
	}

	statusCodeSection := lipgloss.JoinVertical(lipgloss.Left,
		statusCodeStyle.Render("Status Code Distribution"),
		lipgloss.NewStyle().Render(fmt.Sprintf("%.2f%% error rate", m.Metrics.ErrorRate)),
		m.renderStatusCodes(),
	)

	return statusCodeSection, nil
}

func (m *Model) renderErrorSection() (string, error) {
	if m.Metrics.Errors == nil {
		return "", fmt.Errorf("errors map is nil")
	}

	errorSection := lipgloss.JoinVertical(lipgloss.Left,
		errorStyle.Render("Error Distribution"),
		lipgloss.NewStyle().Render(fmt.Sprintf("%.2f%% error rate", m.Metrics.ErrorRate)),
		m.renderErrors(),
	)

	return errorSection, nil
}

func (m *Model) renderLineCharts() (string, error) {
	if m.LatencyGraph == nil || m.ThroughputGraph == nil {
		return "", fmt.Errorf("line charts not initialized")
	}

	lineCharts := lipgloss.JoinHorizontal(lipgloss.Top,
		m.LatencyGraph.View(),
		m.ThroughputGraph.View(),
	)

	return lineCharts, nil
}

func (m Model) runLoadTest() tea.Cmd {
	log.Println("Starting load test...")
	return func() tea.Msg {
		progressCh := make(chan float64)
		go func() {
			for progress := range progressCh {
				m.updateProgress(progress)
			}
		}()

		metrics, err := loadtest.Run(m.Endpoint, m.Duration, m.Concurrency, progressCh)
		return loadTestResultMsg{metrics: metrics, err: err}
	}
}

func (m *Model) updateProgress(progress float64) {
	m.Progress = progress
	m.ProgressBar.SetPercent(progress)
}

func (m *Model) updateLineCharts() {
	log.Println("Updating line charts")
	log.Println("Latency Data:", m.Metrics.LatencyData)
	log.Println("Throughput Data:", m.Metrics.ThroughputData)
	m.LatencyGraph.Data = m.Metrics.LatencyData
	m.ThroughputGraph.Data = m.Metrics.ThroughputData
}

func (m *Model) renderStatusCodes() string {
	log.Printf("Rendering status codes: %v", m.Metrics.StatusCodes)
	if len(m.Metrics.StatusCodes) == 0 {
		return "No status code data available"
	}

	var lines []string
	for code, count := range m.Metrics.StatusCodes {
		lines = append(lines, fmt.Sprintf("[%d] %d responses", code, count))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *Model) renderErrors() string {
	log.Printf("Rendering errors: %v", m.Metrics.Errors)
	if len(m.Metrics.Errors) == 0 {
		return "No error data available"
	}

	var lines []string
	for err, count := range m.Metrics.Errors {
		lines = append(lines, fmt.Sprintf("%s: %d", err, count))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type LineChart struct {
	Title  string
	Data   []float64
	Width  int
	Height int
}

func (c *LineChart) View() string {
	log.Printf("LineChart Data: %v", c.Data)
	log.Printf("LineChart Width: %d", c.Width)
	log.Printf("LineChart Height: %d", c.Height)

	if len(c.Data) == 0 {
		return "No data available"
	}

	if c.Width <= 0 || c.Height <= 0 {
		return "Invalid chart dimensions"
	}

	lc := widgets.NewPlot()
	lc.Title = c.Title
	lc.Data = make([][]float64, 1)
	lc.Data[0] = c.Data
	lc.SetRect(0, 0, c.Width, c.Height)
	lc.AxesColor = ui.ColorWhite
	lc.LineColors[0] = ui.ColorGreen

	buffer := ui.NewBuffer(image.Rect(0, 0, c.Width, c.Height))

	lc.Draw(buffer)

	return buffer.String()
}

func NewProgram(endpoint string, duration time.Duration, concurrency int) *tea.Program {
	return tea.NewProgram(InitialModel(endpoint, duration, concurrency))
}
