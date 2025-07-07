// Package ui provides UI components and formatting utilities
//
// This package provides UI components for displaying bandwidth statistics
// and various formatting utilities for human-readable display of
// bandwidth, bytes, and duration values.
package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Reset       key.Binding
	Pause       key.Binding
	Stats       key.Binding
	DisplayMode key.Binding
	Quit        key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset chart"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p", " "),
			key.WithHelp("p/space", "pause/resume"),
		),
		Stats: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle statusbar"),
		),
		DisplayMode: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "toggle display mode"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// Stats represents various statistics about the monitoring
type Stats struct {
	TotalUpload   uint64
	TotalDownload uint64
	PeakUpload    uint64
	PeakDownload  uint64
	StartTime     time.Time
	// Optimization: cache update interval to reduce repeated calculations
	updateInterval time.Duration
}

// NewStats creates a new stats tracker
func NewStats() *Stats {
	return &Stats{
		StartTime:      time.Now(),
		updateInterval: 500 * time.Millisecond, // Cache the update interval
	}
}

// Update updates the statistics
func (s *Stats) Update(upload, download uint64) {
	// Calculate totals based on rate * time
	// upload and download are in bytes per second, so multiply by time interval
	bytesUploadedThisInterval := float64(upload) * s.updateInterval.Seconds()
	bytesDownloadedThisInterval := float64(download) * s.updateInterval.Seconds()

	s.TotalUpload += uint64(bytesUploadedThisInterval)
	s.TotalDownload += uint64(bytesDownloadedThisInterval)

	// Update peak values
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

// Reset resets all statistics
func (s *Stats) Reset() {
	s.TotalUpload = 0
	s.TotalDownload = 0
	s.PeakUpload = 0
	s.PeakDownload = 0
	s.StartTime = time.Now()
}

// Enhanced UI components
type Components struct {
	stats *Stats
}

// NewComponents creates new UI components
func NewComponents() *Components {
	return &Components{
		stats: NewStats(),
	}
}

// GetStats returns the current statistics
func (c *Components) GetStats() *Stats {
	return c.stats
}

// FormatBandwidth formats bandwidth for UI display
func FormatBandwidth(bps uint64) string {
	const unit = 1024
	if bps < unit {
		return fmt.Sprintf("%d B/s", bps)
	}
	div, exp := uint64(unit), 0
	for n := bps / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	// Optimization: pre-defined units array to avoid string indexing
	units := []string{"KB/s", "MB/s", "GB/s", "TB/s", "PB/s", "EB/s"}
	return fmt.Sprintf("%.2f %s", float64(bps)/float64(div), units[exp])
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		return fmt.Sprintf("%dm%ds", minutes, remainingSeconds)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
}

// FormatBytes formats bytes in a human-readable way
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	// Optimization: avoid multiple comparisons, use single switch
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
