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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"

	"github.com/marcodenic/peaks/internal/chart"
	"github.com/marcodenic/peaks/internal/monitor"
	"github.com/marcodenic/peaks/internal/ui"
)

// getVersion returns the version of the application
func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "dev"
}

// version is set at build time via -ldflags or detected automatically
var version = getVersion()

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
			// Recalculate chart height (same logic as WindowSizeMsg)
			chartHeight := m.height - 1 // Leave room for help text
			if m.showStatusbar {
				chartHeight -= 1 // Leave room for statusbar
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
		title := titleStyle.Render("  🏔️ PEAKS " + version)
		
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

// runCompactMode runs the bandwidth monitor in compact mode (2-line header)
// This forks to background and sets up scroll regions
func runCompactMode(overlay bool, timeMinutes int, size int) {
	// Check if we're already the background daemon
	isDaemon := os.Getenv("PEAKS_DAEMON") == "1"
	
	if !isDaemon {
		// We're the parent - fork to background
		env := append(os.Environ(), "PEAKS_DAEMON=1")
		
		// Build command with flags
		args := []string{"--compact"}
		if overlay {
			args = append(args, "--overlay")
		}
		if timeMinutes != 1 {
			args = append(args, "--time", fmt.Sprintf("%d", timeMinutes))
		}
		if size != 2 {
			args = append(args, "--size", fmt.Sprintf("%d", size))
		}
		
		cmd := exec.Command(os.Args[0], args...)
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		// Start the daemon
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
		
		// Give daemon a moment to start
		time.Sleep(50 * time.Millisecond)
		
		// Move cursor up and clear the command line that was just printed
		// We need to clear the "./peaks --compact" line that the shell echoed
		fmt.Print("\033[1A")                          // Move up 1 line (to where the command was)
		fmt.Print("\033[2K")                          // Clear that line
		fmt.Print("\r")                               // Return to start of line
		
		// Now set up the display properly
		termHeight := getTerminalHeight()
		fmt.Print("\033[2J")                          // Clear entire screen
		fmt.Print("\033[H")                           // Move to home
		
		// Reserve top N lines based on size
		for i := 0; i < size; i++ {
			fmt.Print("\n")
		}
		fmt.Printf("\033[%d;%dr", size+1, termHeight)  // Set scroll region from line (size+1) to bottom
		fmt.Printf("\033[%d;1H", size+1)               // Move to line (size+1), column 1
		
		// Save PID for cleanup (user can find it with: pgrep peaks)
		pidFile := fmt.Sprintf("/tmp/peaks-%d.pid", os.Getpid())
		os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
		
		// Parent exits, returns control to shell
		return
	}
	
	// We're the daemon - do the actual monitoring
	runCompactDaemon(overlay, timeMinutes, size)
}

// runCompactDaemon runs as a background daemon
func runCompactDaemon(overlay bool, timeMinutes int, size int) {
	// Initialize monitor and chart
	mon := monitor.NewBandwidthMonitor()
	ch := chart.NewBrailleChart(defaultDataPoints)
	
	// Set overlay mode if requested
	ch.SetOverlayMode(overlay)
	
	// Map time minutes to TimeScale
	var timeScale chart.TimeScale
	switch timeMinutes {
	case 1:
		timeScale = chart.TimeScale1Min
	case 5:
		timeScale = chart.TimeScale5Min
	case 10:
		timeScale = chart.TimeScale10Min
	case 30:
		timeScale = chart.TimeScale30Min
	case 60:
		timeScale = chart.TimeScale60Min
	default:
		timeScale = chart.TimeScale1Min
	}
	ch.SetTimeScale(timeScale)
	
	// Store enough data for the requested time window
	// 2 points per second * 60 seconds * minutes
	maxDataPoints := 2 * 60 * timeMinutes
	if maxDataPoints < defaultDataPoints {
		maxDataPoints = defaultDataPoints
	}
	ch.SetMaxPoints(maxDataPoints)

	// Get initial terminal dimensions
	termWidth := getTerminalWidth()
	termHeight := getTerminalHeight()

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	defer func() {
		// Cleanup: restore normal scroll region and clear top lines
		fmt.Printf("\033[1;%dr", termHeight)      // Reset scroll region to full screen
		for i := 1; i <= size; i++ {
			fmt.Printf("\033[%d;1H\033[2K", i)    // Clear each line
		}
		fmt.Print("\033[2J\033[H")                // Clear screen and move home
	}()

	// Main update loop
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get current bandwidth
			upload, download, err := mon.GetCurrentRates()
			if err == nil {
				ch.AddDataPoint(upload, download)
			}

			// Check for terminal resize
			newWidth := getTerminalWidth()
			newHeight := getTerminalHeight()
			if newWidth != termWidth || newHeight != termHeight {
				termWidth = newWidth
				termHeight = newHeight
			}

			// Render compact chart with current terminal width and size
			compactView := ch.RenderCompactWithSize(termWidth, size)

			// Update top N lines WITHOUT affecting scroll region or cursor
			fmt.Print("\0337")                    // Save cursor position
			
			// Clear and update each line to prevent wrapping/leftover chars
			lines := strings.Split(compactView, "\n")
			for i := 0; i < size && i < len(lines); i++ {
				fmt.Printf("\033[%d;1H\033[2K", i+1) // Move to line i+1 and clear entire line
				fmt.Print(lines[i])                   // Draw the line
			}
			
			fmt.Print("\0338")                    // Restore cursor position

		case <-sigChan:
			return
		}
	}
}

// getTerminalHeight gets terminal height
func getTerminalHeight() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	
	ws := &winsize{}
	
	// Try stdout first (works better in daemon mode)
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	// If stdout fails, try stderr
	if int(retCode) == -1 {
		retCode, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stderr),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}
	
	// If both fail, try stdin as last resort
	if int(retCode) == -1 {
		retCode, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}

	if int(retCode) == -1 {
		return 24 // Fallback
	}
	
	return int(ws.Row)
}

// getTerminalWidth attempts to get terminal width using ioctl
func getTerminalWidth() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	
	ws := &winsize{}
	
	// Try stdout first (works better in daemon mode)
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	// If stdout fails, try stderr
	if int(retCode) == -1 {
		retCode, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stderr),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}
	
	// If both fail, try stdin as last resort
	if int(retCode) == -1 {
		retCode, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(ws)))
	}

	if int(retCode) == -1 {
		return 80 // Fallback
	}
	
	return int(ws.Col)
}

func main() {
	// Parse command-line flags
	compactMode := flag.Bool("compact", false, "run in compact mode (2-line display at top of terminal)")
	compactOverlay := flag.Bool("overlay", false, "use overlay mode in compact view (both bars from bottom)")
	compactTime := flag.Int("time", 1, "time window in minutes for compact mode (1, 5, 10, 30, 60)")
	compactSize := flag.Int("size", 2, "height in lines for compact mode (2, 3, 4, etc.)")
	showVersion := flag.Bool("version", false, "show version information")
	flag.BoolVar(showVersion, "v", false, "show version information (shorthand)")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("PEAKS %s\n", getVersion())
		return
	}

	// Run in compact mode or full mode
	if *compactMode {
		runCompactMode(*compactOverlay, *compactTime, *compactSize)
	} else {
		p := tea.NewProgram(
			initialModel(),
			tea.WithAltScreen(),
		)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v", err)
		}
	}
}

