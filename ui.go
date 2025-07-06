package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Reset key.Binding
	Pause key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Reset, k.Pause, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Reset, k.Pause, k.Help, k.Quit},
	}
}

var keys = KeyMap{
	Reset: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reset chart"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p", " "),
		key.WithHelp("p/space", "pause/resume"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Stats represents various statistics about the monitoring
type Stats struct {
	TotalUpload   uint64
	TotalDownload uint64
	PeakUpload    uint64
	PeakDownload  uint64
	StartTime     time.Time
}

// NewStats creates a new stats tracker
func NewStats() *Stats {
	return &Stats{
		StartTime: time.Now(),
	}
}

// Update updates the statistics
func (s *Stats) Update(upload, download uint64) {
	s.TotalUpload += upload
	s.TotalDownload += download

	if upload > s.PeakUpload {
		s.PeakUpload = upload
	}
	if download > s.PeakDownload {
		s.PeakDownload = download
	}
}

// GetUptime returns the uptime duration
func (s *Stats) GetUptime() time.Duration {
	return time.Since(s.StartTime)
}

// Enhanced UI components
type UIComponents struct {
	help  help.Model
	stats *Stats
}

// NewUIComponents creates new UI components
func NewUIComponents() *UIComponents {
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF"))
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF"))
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C"))

	return &UIComponents{
		help:  h,
		stats: NewStats(),
	}
}

// RenderStats creates a beautiful stats display
func (ui *UIComponents) RenderStats(upload, download uint64) string {
	ui.stats.Update(upload, download)

	statStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB")).
		Background(lipgloss.Color("#1F2937")).
		Padding(1, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F9FAFB")).
		Bold(true)

	uptime := ui.stats.GetUptime()

	lines := []string{
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Uptime: "),
			valueStyle.Render(formatDuration(uptime)),
		),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Peak ↑: "),
			valueStyle.Render(formatBandwidth(ui.stats.PeakUpload)),
		),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Peak ↓: "),
			valueStyle.Render(formatBandwidth(ui.stats.PeakDownload)),
		),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Total ↑: "),
			valueStyle.Render(formatBytes(ui.stats.TotalUpload)),
		),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Total ↓: "),
			valueStyle.Render(formatBytes(ui.stats.TotalDownload)),
		),
	}

	return statStyle.Render(strings.Join(lines, "\n"))
}

// RenderHelp creates a beautiful help display
func (ui *UIComponents) RenderHelp() string {
	return ui.help.View(keys)
}

// formatBandwidth formats bandwidth for UI display
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

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - hours*60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// formatBytes formats bytes in a human-readable way
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
