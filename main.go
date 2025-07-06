package main

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Update frequency for bandwidth monitoring
	updateInterval = 500 * time.Millisecond
	// Maximum data points to keep for the chart
	maxDataPoints = 120 // 60 seconds of history at 2 FPS
)

var (
	// Global styles using the latest Lip Gloss features
	uploadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")). // Red for upload
			Bold(true)

	downloadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34D399")). // Green for download
			Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB")).
			Background(lipgloss.Color("#374151")).
			Padding(0, 1).
			MarginTop(1)
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(updateInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	monitor   *BandwidthMonitor
	chart     *BrailleChart
	ui        *UIComponents
	width     int
	height    int
	ready     bool
	quitting  bool
	paused    bool
	showHelp  bool
	showStats bool
	lastTick  time.Time
}

func initialModel() model {
	monitor := NewBandwidthMonitor()
	chart := NewBrailleChart(maxDataPoints)
	ui := NewUIComponents()

	return model{
		monitor:   monitor,
		chart:     chart,
		ui:        ui,
		lastTick:  time.Now(),
		showStats: false, // Start with stats disabled
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Set chart width to fit terminal width with minimal padding
		m.chart.SetWidth(msg.Width - 2) // Account for minimal padding
		// Set chart height to fill most of the terminal height - be more aggressive
		m.chart.SetHeight(msg.Height - 2) // Account for footer and help text (reduced from 4 to 2)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.Reset):
			m.chart.Reset()
			m.ui.stats = NewStats()
			return m, nil
		case key.Matches(msg, keys.Pause):
			m.paused = !m.paused
			return m, nil
		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case msg.String() == "s":
			m.showStats = !m.showStats
			return m, nil
		}

	case tickMsg:
		if !m.paused {
			// Update bandwidth data
			upload, download, err := m.monitor.GetCurrentRates()
			if err == nil {
				m.chart.AddDataPoint(upload, download)
				m.ui.stats.Update(upload, download)
			}
		}
		m.lastTick = time.Time(msg)
		return m, tickCmd()

	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Initializing beautiful bandwidth monitor..."
	}

	if m.quitting {
		goodbyeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true).
			Padding(1)
		return goodbyeStyle.Render("Thanks for using Peaks! 🏔️") + "\n"
	}

	// Get current rates for display
	upload, download, _ := m.monitor.GetCurrentRates()

	// Create the chart with full height
	chartContent := m.chart.Render()

	// Create the main display area (just the chart)
	mainContent := chartContent

	// Create the footer with current stats
	uploadText := uploadStyle.Render(fmt.Sprintf("↑ %s", formatBandwidth(upload)))
	downloadText := downloadStyle.Render(fmt.Sprintf("↓ %s", formatBandwidth(download)))

	footer := footerStyle.Render(fmt.Sprintf("%s  %s", uploadText, downloadText))

	// Main content is just the chart
	contentArea := mainContent

	// Add footer
	contentWithFooter := lipgloss.JoinVertical(
		lipgloss.Left,
		contentArea,
		footer,
	)

	// Add help if shown
	if m.showHelp {
		helpContent := m.ui.RenderHelp()
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Background(lipgloss.Color("#1F2937")).
			Padding(1).
			Margin(1, 0)

		styledHelp := helpStyle.Render(helpContent)
		contentWithFooter = lipgloss.JoinVertical(
			lipgloss.Left,
			contentWithFooter,
			styledHelp,
		)
	} else {
		// Show mini help
		miniHelpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Faint(true)

		miniHelp := miniHelpStyle.Render("Press '?' for help • 'p' to pause • 's' to toggle stats • 'r' to reset • 'q' to quit")
		contentWithFooter = lipgloss.JoinVertical(
			lipgloss.Left,
			contentWithFooter,
			miniHelp,
		)
	}

	// Return the content without centering to fill the terminal
	return contentWithFooter
}

func main() {
	// Create a beautiful app with enhanced features
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(15), // Smooth but efficient
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func formatBandwidth(bps uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bps >= GB:
		return fmt.Sprintf("%.2f GB/s", float64(bps)/GB)
	case bps >= MB:
		return fmt.Sprintf("%.2f MB/s", float64(bps)/MB)
	case bps >= KB:
		return fmt.Sprintf("%.2f KB/s", float64(bps)/KB)
	default:
		return fmt.Sprintf("%d B/s", bps)
	}
}
