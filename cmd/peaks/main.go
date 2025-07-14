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
//	m:        Toggle display mode (split/overlay)
//	l:        Cycle scaling mode (linear/logarithmic/square root)
//	t:        Cycle time scale (1/3/5/10/15/30/60 minutes)
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

// version is set at build time via -ldflags
var version = "dev"

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
	chart := chart.NewBrailleChart(defaultDataPoints)
	// Always store 60 minutes of data to support any time scale
	maxDataPoints := 60 * 60 * 2 // 60 minutes * 60 seconds * 2 points per second  
	chart.SetMaxPoints(maxDataPoints)
	
	m := model{
		monitor: monitor.NewBandwidthMonitor(),
		chart:   chart,
		ui:      ui.NewComponents(),
		keys:    ui.DefaultKeyMap(),
	}

	// Create statusbar with 4 sections - no background colors to avoid conflicts with styled text
	m.statusbar = statusbar.New(
		// Current rates section
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#E5E7EB", Light: "#1F2937"},
		},
		// Peak values section
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#9CA3AF", Light: "#6B7280"},
		},
		// Totals section
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#9CA3AF"},
		},
		// Uptime section
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#60A5FA", Light: "#2563EB"},
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

		// Always store 60 minutes of data (regardless of selected time scale)
		// This ensures we have enough data for any time scale selection
		maxDataPoints := 60 * 60 * 2 // 60 minutes * 60 seconds * 2 points per second
		m.chart.SetMaxPoints(maxDataPoints)

		// Update chart dimensions (always responsive to terminal width)
		// Account for: help text (1 line) + status bar (1 line if shown)
		chartHeight := m.height - 1 // Leave room for help text
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
			chartHeight := m.height - 3 // Help text + buffer + extra safety
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

		case key.Matches(msg, m.keys.TimeScale):
			// Cycle through time scales
			m.chart.CycleTimeScale()
			// No need to change max points - we always store 60 minutes of data
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

	// Define arrow colors to match upload/download data
	uploadArrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#EF4444", Light: "#DC2626"}) // Red for upload
	downloadArrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#10B981", Light: "#047857"}) // Green for download

	// Define value colors - current rates are more opaque (prominent)
	currentUploadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#EF4444", Light: "#DC2626"}) // Full opacity red
	currentDownloadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#10B981", Light: "#047857"}) // Full opacity green

	// Peak values - semi-transparent
	peakUploadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#DC2626", Light: "#991B1B"}) // Muted red
	peakDownloadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#059669", Light: "#065F46"}) // Muted green

	// Total values - same opacity as peak values
	totalUploadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#DC2626", Light: "#991B1B"}) // Same muted red as peaks
	totalDownloadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#059669", Light: "#065F46"}) // Same muted green as peaks

	// Format current rates with colored arrows and values
	uploadFormatted := ui.FormatBandwidth(m.currentUpload)
	downloadFormatted := ui.FormatBandwidth(m.currentDownload)
	currentRates := fmt.Sprintf("%s%s %s%s", 
		downloadArrowStyle.Render("↓"), currentDownloadStyle.Render(fmt.Sprintf("%11s", downloadFormatted)),
		uploadArrowStyle.Render("↑"), currentUploadStyle.Render(fmt.Sprintf("%11s", uploadFormatted)))

	// Format peak values with colored arrows and values
	peakUploadFormatted := ui.FormatBandwidth(stats.PeakUpload)
	peakDownloadFormatted := ui.FormatBandwidth(stats.PeakDownload)
	peakValues := fmt.Sprintf("Peak: %s %s %s %s", 
		downloadArrowStyle.Render("↓"), peakDownloadStyle.Render(fmt.Sprintf("%9s", peakDownloadFormatted)),
		uploadArrowStyle.Render("↑"), peakUploadStyle.Render(fmt.Sprintf("%9s", peakUploadFormatted)))

	// Format totals with colored arrows and values
	totalUploadFormatted := ui.FormatBytes(stats.TotalUpload)
	totalDownloadFormatted := ui.FormatBytes(stats.TotalDownload)
	totalValues := fmt.Sprintf("Total: %s %s %s %s", 
		downloadArrowStyle.Render("↓"), totalDownloadStyle.Render(fmt.Sprintf("%8s", totalDownloadFormatted)),
		uploadArrowStyle.Render("↑"), totalUploadStyle.Render(fmt.Sprintf("%8s", totalUploadFormatted)))

	// Format uptime and display mode and scaling mode and time scale
	uptimeValue := fmt.Sprintf("Up: %s | Mode: %s | Scale: %s | Time: %s",
		ui.FormatDuration(stats.GetUptime()),
		m.displayMode,
		m.chart.GetScalingModeName(),
		m.chart.GetTimeScaleName())

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

	// Chart
	chartView := m.chart.Render()
	view.WriteString(chartView)

	// Statusbar
	if m.showStatusbar {
		view.WriteString("\n")
		view.WriteString(m.statusbar.View())
	}

	// Title and controls help
	if m.height > 10 { // Only show if we have enough space
		view.WriteString("\n")
		
		// Create title
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			Bold(true)
		title := titleStyle.Render("🏔️ PEAKS " + version)
		
		// Create help text
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
		controls := "r: reset • p: pause • s: statusbar • m: mode • l: scaling • t: time • q: quit"
		if m.paused {
			controls = "r: reset • p: resume • s: statusbar • m: mode • l: scaling • t: time • q: quit"
		}
		help := helpStyle.Render(controls)
		
		// Calculate spacing to right-align help
		titleWidth := lipgloss.Width(title)
		helpWidth := lipgloss.Width(help)
		availableWidth := m.width
		
		if titleWidth + helpWidth < availableWidth {
			// Right-align help text
			spacingWidth := availableWidth - titleWidth - helpWidth
			spacing := strings.Repeat(" ", spacingWidth)
			bottomLine := title + spacing + help
			view.WriteString(bottomLine)
		} else {
			// Fall back to just showing title if not enough space
			view.WriteString(title)
		}
	}

	// Ensure we don't end with trailing newlines
	result := view.String()
	result = strings.TrimRight(result, "\n\r ")
	return result
}

func main() {
	p := tea.NewProgram(
		initialModel(), 
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}

