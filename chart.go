// Package main - braille chart rendering functionality
//
// This file implements high-resolution braille chart rendering for terminal displays.
// It creates split-axis charts with upload data below and download data above a
// central axis, using Unicode braille characters for detailed visualization.
package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// Chart configuration constants
	minChartHeight = 8                 // Minimum chart height in rows
	brailleDots    = 4                 // Braille has 4 vertical dots per character
	brailleBase    = 0x2800            // Base braille character code
	maxScaleLimit  = 100 * 1024 * 1024 // 100MB/s maximum scale
)

var (
	// Chart-specific styles for braille characters
	uploadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")). // Red for upload
			Bold(true)

	downloadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34D399")). // Green for download
			Bold(true)
)

// BrailleChart creates beautiful braille-based charts for terminal display
type BrailleChart struct {
	width        int
	height       int
	maxPoints    int
	uploadData   []uint64
	downloadData []uint64
	maxValue     uint64
	minHeight    int
	// Optimization: track current max without full recalculation
	currentMax   uint64
	// Optimization: pre-allocated string builder for rendering
	builder      strings.Builder
	// Optimization: pre-allocated slice for lines to avoid repeated allocations
	lines        []strings.Builder
}

// DataPoint represents a single measurement point
type DataPoint struct {
	Upload   uint64
	Download uint64
}

// NewBrailleChart creates a new braille chart
func NewBrailleChart(maxPoints int) *BrailleChart {
	return &BrailleChart{
		width:        80,
		height:       20,
		maxPoints:    maxPoints,
		// Optimization: pre-allocate slices with fixed capacity to avoid reallocations
		uploadData:   make([]uint64, 0, maxPoints),
		downloadData: make([]uint64, 0, maxPoints),
		maxValue:     1024, // Start with 1KB minimum scale
		minHeight:    minChartHeight,
		currentMax:   0,
		// Optimization: pre-allocate string builders
		lines:        make([]strings.Builder, 0, 50), // Pre-allocate for typical chart heights
	}
}

// SetWidth sets the chart width
func (bc *BrailleChart) SetWidth(width int) {
	bc.width = width
	if bc.width < 20 {
		bc.width = 20
	}
}

// SetHeight sets the chart height
func (bc *BrailleChart) SetHeight(height int) {
	bc.height = height
	if bc.height < bc.minHeight {
		bc.height = bc.minHeight
	}
}

// AddDataPoint adds a new data point to the chart
func (bc *BrailleChart) AddDataPoint(upload, download uint64) {
	// Optimization: track current max efficiently
	if upload > bc.currentMax {
		bc.currentMax = upload
	}
	if download > bc.currentMax {
		bc.currentMax = download
	}

	// Add new data points
	bc.uploadData = append(bc.uploadData, upload)
	bc.downloadData = append(bc.downloadData, download)

	// Remove old data points if we exceed max
	if len(bc.uploadData) > bc.maxPoints {
		// Optimization: check if we're removing the max value
		removedUpload := bc.uploadData[0]
		bc.uploadData = bc.uploadData[1:]
		
		// If we removed the max value, we need to recalculate
		if removedUpload == bc.currentMax {
			bc.recalculateMax()
		}
	}
	if len(bc.downloadData) > bc.maxPoints {
		// Optimization: check if we're removing the max value
		removedDownload := bc.downloadData[0]
		bc.downloadData = bc.downloadData[1:]
		
		// If we removed the max value, we need to recalculate
		if removedDownload == bc.currentMax {
			bc.recalculateMax()
		}
	}

	// Update max value for scaling
	bc.updateMaxValue()
}

// recalculateMax recalculates the current max when the previous max was removed
func (bc *BrailleChart) recalculateMax() {
	bc.currentMax = 0
	for _, val := range bc.uploadData {
		if val > bc.currentMax {
			bc.currentMax = val
		}
	}
	for _, val := range bc.downloadData {
		if val > bc.currentMax {
			bc.currentMax = val
		}
	}
}

// updateMaxValue updates the maximum value for scaling
func (bc *BrailleChart) updateMaxValue() {
	// Optimization: use tracked currentMax instead of recalculating every time
	currentMax := bc.currentMax

	// Use more aggressive scaling that utilizes full height
	if currentMax < 1024 {
		bc.maxValue = 1024
	} else if currentMax > bc.maxValue {
		// Scale up with only 10% headroom instead of 100%
		bc.maxValue = currentMax + (currentMax / 10)
	} else if currentMax < bc.maxValue/2 && bc.maxValue > 1024 {
		// Scale down more aggressively when current max is less than half
		bc.maxValue = currentMax + (currentMax / 10)
	}

	// Ensure minimum headroom
	if bc.maxValue < currentMax {
		bc.maxValue = currentMax
	}

	// Reasonable bounds
	if bc.maxValue > maxScaleLimit {
		bc.maxValue = maxScaleLimit
	}
}

// Reset clears all data
func (bc *BrailleChart) Reset() {
	// Optimization: reuse existing slice capacity instead of creating new ones
	bc.uploadData = bc.uploadData[:0]
	bc.downloadData = bc.downloadData[:0]
	bc.maxValue = 1024
	bc.currentMax = 0
	// Reset string builders
	bc.builder.Reset()
	for i := range bc.lines {
		bc.lines[i].Reset()
	}
}

// Render creates the beautiful braille chart with split axis
func (bc *BrailleChart) Render() string {
	if len(bc.uploadData) == 0 && len(bc.downloadData) == 0 {
		return bc.renderEmptyChart()
	}

	// Calculate dimensions - use the full available width
	chartWidth := bc.width
	if chartWidth < 20 {
		chartWidth = 20
	}

	// Create the chart
	chart := bc.renderBrailleChart(chartWidth)

	// Return just the chart without legend, axis, or scale
	return chart
}

// renderEmptyChart renders a placeholder when no data is available
func (bc *BrailleChart) renderEmptyChart() string {
	// Just return empty space - don't show anything until data arrives
	chartHeight := bc.height - 3
	if chartHeight < minChartHeight {
		chartHeight = minChartHeight
	}

	// Return empty lines to fill the space
	var lines []string
	for i := 0; i < chartHeight; i++ {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderBrailleChart creates the actual braille chart with split axis
func (bc *BrailleChart) renderBrailleChart(width int) string {
	if len(bc.uploadData) == 0 && len(bc.downloadData) == 0 {
		return bc.renderEmptyChart()
	}

	// Chart dimensions - use the dynamic height more aggressively
	chartHeight := bc.height - 3 // Reserve space for footer and help, but use most of the height
	if chartHeight < minChartHeight {
		chartHeight = minChartHeight
	}
	dotsPerLine := brailleDots               // Braille has 4 vertical dots per character
	totalHeight := chartHeight * dotsPerLine // Total vertical resolution
	halfHeight := totalHeight / 2            // Split point for upload/download

	// Get the data to display - always show the most recent data on the right
	dataLen := len(bc.uploadData)
	if len(bc.downloadData) > dataLen {
		dataLen = len(bc.downloadData)
	}

	// Optimization: ensure we have enough pre-allocated string builders
	if len(bc.lines) < chartHeight {
		for i := len(bc.lines); i < chartHeight; i++ {
			bc.lines = append(bc.lines, strings.Builder{})
		}
	}

	// Optimization: reset and reuse existing builders instead of creating new strings
	for i := 0; i < chartHeight; i++ {
		bc.lines[i].Reset()
		bc.lines[i].Grow(width) // Pre-allocate capacity
	}

	// Optimization: pre-calculate values to avoid repeated calculations
	maxValueFloat := float64(bc.maxValue)
	halfHeightFloat := float64(halfHeight)

	for col := 0; col < width; col++ {
		// Calculate which data point this column represents
		// New data should always appear on the right side
		dataIdx := dataLen - (width - col)

		var uploadVal, downloadVal uint64
		if dataIdx >= 0 && dataIdx < len(bc.uploadData) {
			uploadVal = bc.uploadData[dataIdx]
		}
		if dataIdx >= 0 && dataIdx < len(bc.downloadData) {
			downloadVal = bc.downloadData[dataIdx]
		}

		// Convert to chart heights (0 to halfHeight for each direction)
		// Optimization: use pre-calculated float values to avoid repeated conversions
		uploadHeight := int(float64(uploadVal) / maxValueFloat * halfHeightFloat)
		downloadHeight := int(float64(downloadVal) / maxValueFloat * halfHeightFloat)

		// Ensure we don't exceed bounds
		if uploadHeight > halfHeight {
			uploadHeight = halfHeight
		}
		if downloadHeight > halfHeight {
			downloadHeight = halfHeight
		}

		// Create braille character for each line
		for line := 0; line < chartHeight; line++ {
			char := bc.createBrailleCharForLineSplit(line, uploadHeight, downloadHeight, halfHeight, dotsPerLine)
			bc.lines[line].WriteString(char)
		}
	}

	// Optimization: use main builder to join lines efficiently
	bc.builder.Reset()
	bc.builder.Grow(chartHeight * (width + 1)) // Pre-allocate capacity
	for i := 0; i < chartHeight; i++ {
		bc.builder.WriteString(bc.lines[i].String())
		if i < chartHeight-1 {
			bc.builder.WriteString("\n")
		}
	}

	return bc.builder.String()
}

// createBrailleCharForLineSplit creates a braille character for a specific line with split axis
func (bc *BrailleChart) createBrailleCharForLineSplit(line, uploadHeight, downloadHeight, halfHeight, dotsPerLine int) string {
	// Optimization: early return for empty characters
	if uploadHeight == 0 && downloadHeight == 0 {
		return " "
	}

	// Base braille character
	base := brailleBase
	var dots int

	hasUpload := false
	hasDownload := false

	// Braille character layout:
	// 0 3
	// 1 4
	// 2 5
	// 6 7

	// Calculate the vertical range of this braille character
	// Line 0 is at the top, line 5 is at the bottom (natural order)
	lineTop := line * dotsPerLine

	// Optimization: pre-calculate dot patterns to avoid repeated switch statements
	dotPatterns := [4]int{
		0x01 | 0x08, // dots 0,3 (row 0)
		0x02 | 0x10, // dots 1,4 (row 1)
		0x04 | 0x20, // dots 2,5 (row 2)
		0x40 | 0x80, // dots 6,7 (row 3)
	}

	// Check each dot position in this braille character (4 dots vertically)
	for dotRow := 0; dotRow < dotsPerLine; dotRow++ {
		// Calculate the absolute dot position in the chart (from top)
		absoluteDotPos := lineTop + dotRow

		// Check if this dot should be filled for download (above axis)
		// Download fills from axis upward
		if absoluteDotPos < halfHeight {
			// We're above the axis - check if within download area
			// Download should fill from (halfHeight - downloadHeight) up to halfHeight
			distanceFromAxis := halfHeight - absoluteDotPos
			if distanceFromAxis <= downloadHeight {
				hasDownload = true
				dots |= dotPatterns[dotRow]
			}
		}

		// Check if this dot should be filled for upload (below axis)
		// Upload fills from axis down to bottom
		if absoluteDotPos >= halfHeight {
			// We're below the axis - check if within upload area
			distanceFromAxis := absoluteDotPos - halfHeight
			if distanceFromAxis < uploadHeight {
				hasUpload = true
				dots |= dotPatterns[dotRow]
			}
		}
	}

	// Optimization: skip character creation if no dots
	if dots == 0 {
		return " "
	}

	// Create the character
	char := rune(base + dots)
	charStr := string(char)

	// Optimization: reduce style applications - apply only when needed
	if hasUpload && hasDownload {
		// This shouldn't happen with split axis, but just in case
		return uploadStyle.Render(charStr)
	} else if hasUpload {
		return uploadStyle.Render(charStr)
	} else if hasDownload {
		return downloadStyle.Render(charStr)
	}

	return charStr
}
