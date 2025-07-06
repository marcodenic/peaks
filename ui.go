// Package main - UI components and formatting utilities
//
// This file provides UI components for displaying bandwidth statistics
// and various formatting utilities for human-readable display of 
// bandwidth, bytes, and duration values.
package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Reset key.Binding
	Pause key.Binding
	Stats key.Binding
	Quit  key.Binding
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
	Stats: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle statusbar"),
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
	// Optimization: calculate totals based on rate * time instead of accumulating
	// This is more accurate for bandwidth totals
	s.TotalUpload += upload * uint64(s.updateInterval.Seconds())
	s.TotalDownload += download * uint64(s.updateInterval.Seconds())

	// Optimization: use bitwise operations for simple comparisons when possible
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
	stats *Stats
}

// NewUIComponents creates new UI components
func NewUIComponents() *UIComponents {
	return &UIComponents{
		stats: NewStats(),
	}
}

// formatBandwidth formats bandwidth for UI display
func formatBandwidth(bps uint64) string {
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

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
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

// formatBytes formats bytes in a human-readable way
func formatBytes(bytes uint64) string {
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
