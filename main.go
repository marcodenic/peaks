package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
)

var (
	// Update frequency for bandwidth monitoring
	updateInterval = 500 * time.Millisecond
	// Maximum data points to keep for the chart
	maxDataPoints = 120 // 60 seconds of history at 2 FPS
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
	statusbar statusbar.Model
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

	// Create statusbar with 4 sections and nice colors
	sb := statusbar.New(
		// Current rates section - white for visibility
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#E5E7EB", Light: "#1F2937"},
			Background: lipgloss.AdaptiveColor{Dark: "#1F2937", Light: "#E5E7EB"},
		},
		// Peak values section - muted
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#9CA3AF", Light: "#6B7280"},
			Background: lipgloss.AdaptiveColor{Dark: "#1F2937", Light: "#E5E7EB"},
		},
		// Totals section - subtle
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#9CA3AF"},
			Background: lipgloss.AdaptiveColor{Dark: "#1F2937", Light: "#E5E7EB"},
		},
		// Uptime section - blue
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#60A5FA", Light: "#2563EB"},
			Background: lipgloss.AdaptiveColor{Dark: "#1F2937", Light: "#E5E7EB"},
		},
	)

	return model{
		monitor:   monitor,
		chart:     chart,
		ui:        ui,
		statusbar: sb,
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
		// Set statusbar width
		m.statusbar.SetSize(msg.Width)
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

	// Create the footer with comprehensive stats using teacup statusbar
	m.ui.stats.Update(upload, download)

	// Format current rates with truly fixed width to prevent jumping
	// NO STYLING HERE - use plain text with fixed width formatting
	uploadFormatted := formatBandwidth(upload)
	downloadFormatted := formatBandwidth(download)

	// Use fixed width formatting WITHOUT any styling to prevent jumping
	currentRates := fmt.Sprintf("↑%11s ↓%11s", uploadFormatted, downloadFormatted)

	// Format peak values with fixed formatting
	peakUploadFormatted := formatBandwidth(m.ui.stats.PeakUpload)
	peakDownloadFormatted := formatBandwidth(m.ui.stats.PeakDownload)
	peakValues := fmt.Sprintf("Peak: ↑%9s ↓%9s", peakUploadFormatted, peakDownloadFormatted)

	// Format totals with fixed formatting
	totalUploadFormatted := formatBytes(m.ui.stats.TotalUpload)
	totalDownloadFormatted := formatBytes(m.ui.stats.TotalDownload)
	totalValues := fmt.Sprintf("Total: ↑%8s ↓%8s", totalUploadFormatted, totalDownloadFormatted)

	// Format uptime
	uptimeValue := "Up: " + formatDuration(m.ui.stats.GetUptime())

	m.statusbar.SetContent(currentRates, peakValues, totalValues, uptimeValue)
	footer := m.statusbar.View()

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
		panic(err)
	}
}
