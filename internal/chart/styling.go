// Package chart provides gradient and styling functionality for braille charts
package chart

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Chart-specific styles for braille characters - base colors
	baseUploadColor   = lipgloss.Color("#F87171") // Red for upload
	baseDownloadColor = lipgloss.Color("#34D399") // Green for download

	// Color gradients for height-based shading (darker at top, lighter at bottom)
	uploadGradient = ColorGradient{
		Steps: []lipgloss.Color{
			lipgloss.Color("#7F1D1D"), // Dark red (top/darkest)
			lipgloss.Color("#B91C1C"), // Medium-dark red
			lipgloss.Color("#DC2626"), // Medium red
			lipgloss.Color("#EF4444"), // Medium-light red
			lipgloss.Color("#F87171"), // Light red
			lipgloss.Color("#FCA5A5"), // Very light red (bottom/lightest)
		},
	}

	downloadGradient = ColorGradient{
		Steps: []lipgloss.Color{
			lipgloss.Color("#064E3B"), // Dark green (top/darkest)
			lipgloss.Color("#047857"), // Medium-dark green
			lipgloss.Color("#059669"), // Medium green
			lipgloss.Color("#10B981"), // Medium-light green
			lipgloss.Color("#34D399"), // Light green
			lipgloss.Color("#6EE7B7"), // Very light green (bottom/lightest)
		},
	}

	// Yellow gradient for overlay overlap areas (dark to light from top to bottom)
	overlapGradient = ColorGradient{
		Steps: []lipgloss.Color{
			lipgloss.Color("#713F12"), // Very dark yellow/brown (top/darkest)
			lipgloss.Color("#92400E"), // Dark yellow
			lipgloss.Color("#B45309"), // Medium-dark yellow
			lipgloss.Color("#D97706"), // Medium yellow
			lipgloss.Color("#F59E0B"), // Medium-light yellow
			lipgloss.Color("#FBBF24"), // Light yellow
			lipgloss.Color("#FCD34D"), // Very light yellow
			lipgloss.Color("#FDE68A"), // Extremely light yellow (bottom/lightest)
		},
	}

	// Overlap style for overlay mode (fallback style)
	overlapStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FCD34D")). // Yellow for overlap
			Bold(true)

	// Optimization: character cache for styled braille characters
	uploadCharCache   = make(map[string]string, 1536) // 6 gradient steps * 256 chars
	downloadCharCache = make(map[string]string, 1536) // 6 gradient steps * 256 chars
	overlapCharCache  = make(map[rune]string, 256)
)

// clampPercent clamps a value to the 0-1 range
func clampPercent(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

// getGradientStepIndex calculates the gradient step index for a given height percentage
func getGradientStepIndex(heightPercent float64, stepCount int) int {
	// Invert the gradient position: 0.0 (bottom) = lightest, 1.0 (top) = darkest
	invertedPercent := 1.0 - clampPercent(heightPercent)
	
	// Map to gradient step
	stepIndex := int(invertedPercent * float64(stepCount-1))
	if stepIndex >= stepCount {
		stepIndex = stepCount - 1
	}
	return stepIndex
}

// getGradientColor returns a color from the gradient based on height percentage
func (bc *BrailleChart) getGradientColor(heightPercent float64, isUpload bool) lipgloss.Color {
	gradient := downloadGradient
	if isUpload {
		gradient = uploadGradient
	}

	// Check if gradient is available
	stepCount := len(gradient.Steps)
	if stepCount == 0 {
		if isUpload {
			return baseUploadColor
		}
		return baseDownloadColor
	}

	// Get gradient step index and return color
	stepIndex := getGradientStepIndex(heightPercent, stepCount)
	return gradient.Steps[stepIndex]
}

// getStyledCharWithGradient returns a styled character with gradient coloring
func (bc *BrailleChart) getStyledCharWithGradient(char rune, heightPercent float64, isUpload bool) string {
	color := bc.getGradientColor(heightPercent, isUpload)

	// Create cache key
	var cacheKey string
	if isUpload {
		cacheKey = fmt.Sprintf("u_%c_%.2f", char, heightPercent)
	} else {
		cacheKey = fmt.Sprintf("d_%c_%.2f", char, heightPercent)
	}

	// Check cache first
	var cache map[string]string
	if isUpload {
		cache = uploadCharCache
	} else {
		cache = downloadCharCache
	}

	if cached, exists := cache[cacheKey]; exists {
		return cached
	}

	// Create styled character
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	styled := style.Render(string(char))

	// Cache the result
	cache[cacheKey] = styled

	return styled
}

// getStyledCharWithOverlapGradient returns a styled character with yellow overlap gradient coloring
func (bc *BrailleChart) getStyledCharWithOverlapGradient(char rune, heightPercent float64) string {
	// Check if gradient is available
	stepCount := len(overlapGradient.Steps)
	if stepCount == 0 {
		return overlapStyle.Render(string(char))
	}

	// Get gradient step index and color
	stepIndex := getGradientStepIndex(heightPercent, stepCount)
	color := overlapGradient.Steps[stepIndex]
	
	// Create and return styled character
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	return style.Render(string(char))
}

// getStyledChar returns a cached styled character or creates and caches it
func (bc *BrailleChart) getStyledChar(char rune, isUpload bool) string {
	// Create basic styling without gradient for legacy support
	var style lipgloss.Style
	if isUpload {
		style = lipgloss.NewStyle().Foreground(baseUploadColor).Bold(true)
	} else {
		style = lipgloss.NewStyle().Foreground(baseDownloadColor).Bold(true)
	}

	return style.Render(string(char))
}

// getStyledCharOverlay returns a cached styled character for overlay mode
func (bc *BrailleChart) getStyledCharOverlay(char rune, mode string) string {
	var style lipgloss.Style

	switch mode {
	case "upload":
		style = lipgloss.NewStyle().Foreground(baseUploadColor).Bold(true)
	case "download":
		style = lipgloss.NewStyle().Foreground(baseDownloadColor).Bold(true)
	case "overlap":
		return overlapStyle.Render(string(char))
	default:
		return string(char)
	}

	return style.Render(string(char))
}
