// Package main implements Peaks, a beautiful terminal-based bandwidth monitoring tool.
//
// Peaks provides real-time network bandwidth monitoring with high-resolution braille charts,
// built using the Charm TUI ecosystem for a modern terminal interface.
//
// Features:
//   - Real-time bandwidth monitoring with split-axis braille charts
//   - Cross-platform support (Linux, macOS, Windows)
//   - Interactive controls for pause, reset, and statistics
//   - Beautiful color-coded interface with traffic separation
//   - Detailed statistics tracking
//
// Usage:
//
//	peaks
//
// Controls:
//
//	q/Ctrl+C: Quit
//	p/Space:  Pause/Resume
//	r:        Reset chart and statistics
//	s:        Toggle statusbar
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"

	"github.com/marcodenic/peaks/internal/chart"
	"github.com/marcodenic/peaks/internal/monitor"
	"github.com/marcodenic/peaks/internal/ui"
)

const (
	// Update frequency for bandwidth monitoring
	updateInterval = 500 * time.Millisecond
	// Default data points for initial chart creation
	defaultDataPoints = 200
)

// calculateMaxDataPoints calculates the optimal number of data points
// to maintain based on terminal width, ensuring the chart always fills
// the available space while maintaining some buffer for window resizing
func calculateMaxDataPoints(terminalWidth int) int {
	if terminalWidth <= 0 {
		return defaultDataPoints
	}
	// Keep enough history to fill the terminal width plus 50% buffer
	// for smooth scrolling and window resizing
	return int(float64(terminalWidth) * 1.5)
}

// tickMsg represents a tick message for updating the display
type tickMsg time.Time

// tickCmd creates a command that sends tick messages at regular intervals
func tickCmd() tea.Cmd {
	return tea.Tick(updateInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// model represents the application state for the Bubble Tea framework
type model struct {
	monitor   *monitor.BandwidthMonitor
	chart     *chart.BrailleChart
	ui        *ui.Components
	keys      ui.KeyMap
	statusbar statusbar.Model
	width     int
	height    int
	ready     bool
	quitting  bool
	paused    bool
	// Optimization: cache current rates to avoid repeated calculations
	currentUpload   uint64
	currentDownload uint64
	// UI state
	showStatusbar bool
	displayMode   string // "split" or "overlay"
}

// initialModel creates and initializes the application model
func initialModel() model {
	m := model{
		monitor: monitor.NewBandwidthMonitor(),
		chart:   chart.NewBrailleChart(defaultDataPoints),
		ui:      ui.NewComponents(),
		keys:    ui.DefaultKeyMap(),
	}

	// Create statusbar with 4 sections and nice colors
	m.statusbar = statusbar.New(
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

	m.showStatusbar = true
	m.displayMode = "split" // Default to split axis mode
	return m
}

// Init initializes the application
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages and updates the application state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Calculate new max data points based on terminal width
		newMaxPoints := calculateMaxDataPoints(m.width)
		m.chart.SetMaxPoints(newMaxPoints)

		// Update chart dimensions
		chartHeight := m.height - 2 // Leave room for title and status
		if m.showStatusbar {
			chartHeight -= 1 // Leave room for statusbar
		}
		if chartHeight < chart.MinChartHeight {
			chartHeight = chart.MinChartHeight
		}

		m.chart.SetWidth(m.width)
		m.chart.SetHeight(chartHeight)

		// Update statusbar width
		m.statusbar.SetSize(m.width)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Pause):
			m.paused = !m.paused

		case key.Matches(msg, m.keys.Reset):
			m.chart.Reset()
			m.ui.GetStats().Reset()

		case key.Matches(msg, m.keys.Stats):
			m.showStatusbar = !m.showStatusbar
			// Recalculate chart height
			chartHeight := m.height - 2
			if m.showStatusbar {
				chartHeight -= 1
			}
			if chartHeight < chart.MinChartHeight {
				chartHeight = chart.MinChartHeight
			}
			m.chart.SetHeight(chartHeight)

		case key.Matches(msg, m.keys.DisplayMode):
			// Toggle display mode
			if m.displayMode == "split" {
				m.displayMode = "overlay"
				m.chart.SetOverlayMode(true)
			} else {
				m.displayMode = "split"
				m.chart.SetOverlayMode(false)
			}

		case key.Matches(msg, m.keys.ScalingMode):
			// Cycle through scaling modes
			m.chart.CycleScalingMode()
		}

	case tickMsg:
		if !m.paused {
			// Get current bandwidth rates
			upload, download, err := m.monitor.GetCurrentRates()
			if err == nil {
				m.currentUpload = upload
				m.currentDownload = download

				// Update chart with new data
				m.chart.AddDataPoint(upload, download)

				// Update statistics
				m.ui.GetStats().Update(upload, download)

				// Update statusbar
				m.updateStatusbar()
			}
		}

		// Schedule next update
		cmd = tickCmd()
	}

	return m, cmd
}

// updateStatusbar updates the statusbar with current statistics
func (m *model) updateStatusbar() {
	stats := m.ui.GetStats()

	// Format current rates with truly fixed width to prevent jumping
	uploadFormatted := ui.FormatBandwidth(m.currentUpload)
	downloadFormatted := ui.FormatBandwidth(m.currentDownload)
	currentRates := fmt.Sprintf("↑%11s ↓%11s", uploadFormatted, downloadFormatted)

	// Format peak values with fixed formatting
	peakUploadFormatted := ui.FormatBandwidth(stats.PeakUpload)
	peakDownloadFormatted := ui.FormatBandwidth(stats.PeakDownload)
	peakValues := fmt.Sprintf("Peak: ↑%9s ↓%9s", peakUploadFormatted, peakDownloadFormatted)

	// Format totals with fixed formatting
	totalUploadFormatted := ui.FormatBytes(stats.TotalUpload)
	totalDownloadFormatted := ui.FormatBytes(stats.TotalDownload)
	totalValues := fmt.Sprintf("Total: ↑%8s ↓%8s", totalUploadFormatted, totalDownloadFormatted)

	// Format uptime
	uptimeValue := "Up: " + ui.FormatDuration(stats.GetUptime())

	m.statusbar.SetContent(currentRates, peakValues, totalValues, uptimeValue)
}

// View renders the application UI
func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.quitting {
		return "\n  Goodbye!\n"
	}

	var view strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#60A5FA")). // Blue
		MarginBottom(1)

	title := titleStyle.Render("Peaks - Bandwidth Monitor")
	view.WriteString(title)
	view.WriteString("\n")

	// Chart
	chartView := m.chart.Render()
	view.WriteString(chartView)

	// Statusbar
	if m.showStatusbar {
		view.WriteString("\n")
		view.WriteString(m.statusbar.View())
	}

	// Controls help
	if m.height > 10 { // Only show if we have enough space
		view.WriteString("\n")
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

		controls := "r: reset • p: pause • s: statusbar • m: mode • q: quit"
		if m.paused {
			controls = "r: reset • p: resume • s: statusbar • m: mode • q: quit"
		}

		view.WriteString(helpStyle.Render(controls))
	}

	return view.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
