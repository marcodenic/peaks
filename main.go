package main

import (
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

	// Create the footer with comprehensive stats
	footer := m.renderStatusBar(upload, download)

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

// renderStatusBar creates a comprehensive status bar with all stats
func (m model) renderStatusBar(upload, download uint64) string {
	// Update stats
	m.ui.stats.Update(upload, download)

	// Define styles for the status bar
	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#E5E7EB")).
		Padding(0, 1).
		MarginTop(1).
		Width(m.width) // Set width to terminal width to ensure full background coverage

	// Current rates with fixed width to prevent layout shift
	currentUploadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F87171")).
		Bold(true).
		Width(10).
		Align(lipgloss.Right)
	
	currentDownloadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#34D399")).
		Bold(true).
		Width(10).
		Align(lipgloss.Right)

	// Peak values with muted styling
	peakStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Width(11).
		Align(lipgloss.Right)

	// Total values with subtle styling
	totalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Width(11).
		Align(lipgloss.Right)

	// Uptime with accent color
	uptimeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#60A5FA")).
		Bold(true)

	// Labels with consistent styling
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D1D5DB")).
		Faint(true)

	// Format values with fixed width
	currentUpload := currentUploadStyle.Render(formatBandwidth(upload))
	currentDownload := currentDownloadStyle.Render(formatBandwidth(download))
	peakUpload := peakStyle.Render(formatBandwidth(m.ui.stats.PeakUpload))
	peakDownload := peakStyle.Render(formatBandwidth(m.ui.stats.PeakDownload))
	totalUpload := totalStyle.Render(formatBytes(m.ui.stats.TotalUpload))
	totalDownload := totalStyle.Render(formatBytes(m.ui.stats.TotalDownload))
	uptime := uptimeStyle.Render(formatDuration(m.ui.stats.GetUptime()))

	// Create more compact sections
	currentSection := lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("↑"),
		currentUpload,
		labelStyle.Render(" ↓"),
		currentDownload,
	)

	peakSection := lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Peak: ↑"),
		peakUpload,
		labelStyle.Render(" ↓"),
		peakDownload,
	)

	totalSection := lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Total: ↑"),
		totalUpload,
		labelStyle.Render(" ↓"),
		totalDownload,
	)

	uptimeSection := lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Up: "),
		uptime,
	)

	// Join all sections with compact separators
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151")).
		Faint(true)

	// Create a more compact layout
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		currentSection,
		separatorStyle.Render(" │ "),
		peakSection,
		separatorStyle.Render(" │ "),
		totalSection,
		separatorStyle.Render(" │ "),
		uptimeSection,
	)

	// Create a content wrapper that fills the full width
	contentWrapper := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Left)

	return barStyle.Render(contentWrapper.Render(statusContent))
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
		panic(err)
	}
}
