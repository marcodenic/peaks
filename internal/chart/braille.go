// Package chart provides braille chart rendering functionality
//
// This package implements high-resolution braille chart rendering for terminal displays.
// It creates split-axis charts with upload data below and download data above a
// central axis, using Unicode braille characters for detailed visualization.
package chart

import (
	"strings"
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
	currentMax uint64
	// Optimization: pre-allocated string builder for rendering
	builder strings.Builder
	// Optimization: pre-allocated slice for lines to avoid repeated allocations
	lines []strings.Builder
	// Display mode: false = split axis, true = overlay mode
	overlayMode bool
	// Scaling mode: how the data is scaled (linear, logarithmic, square root)
	scalingMode ScalingMode
	// Time scale: the time window for data display
	timeScale TimeScale
}

// NewBrailleChart creates a new braille chart
func NewBrailleChart(maxPoints int) *BrailleChart {
	return &BrailleChart{
		width:     defaultChartWidth,
		height:    defaultChartHeight,
		maxPoints: maxPoints,
		// Optimization: pre-allocate slices with fixed capacity to avoid reallocations
		uploadData:   make([]uint64, 0, maxPoints),
		downloadData: make([]uint64, 0, maxPoints),
		maxValue:     1024, // Start with 1KB minimum scale
		minHeight:    MinChartHeight,
		currentMax:   0,
		// Optimization: pre-allocate string builders
		lines:       make([]strings.Builder, 0, defaultMaxPoints), // Pre-allocate for typical chart heights
		overlayMode: false,                                        // Default to split axis mode
		scalingMode: ScalingLogarithmic,                          // Default to logarithmic scaling
		timeScale:   TimeScale1Min,                               // Default to 1 minute time scale
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

// SetOverlayMode sets the display mode
func (bc *BrailleChart) SetOverlayMode(enabled bool) {
	bc.overlayMode = enabled
}

// ToggleOverlayMode toggles between split axis and overlay mode
func (bc *BrailleChart) ToggleOverlayMode() {
	bc.overlayMode = !bc.overlayMode
}

// IsOverlayMode returns true if overlay mode is enabled
func (bc *BrailleChart) IsOverlayMode() bool {
	return bc.overlayMode
}

// GetWidth returns the chart width
func (bc *BrailleChart) GetWidth() int {
	return bc.width
}

// Render renders the braille chart as a string
func (bc *BrailleChart) Render() string {
	if len(bc.uploadData) == 0 && len(bc.downloadData) == 0 {
		return bc.renderEmptyChart()
	}

	// Update scaling based on currently visible data before rendering
	bc.updateMaxValue()

	// Reset and prepare string builder
	bc.builder.Reset()
	bc.builder.Grow(bc.width * bc.height * 4) // Estimate capacity

	// Ensure we have enough pre-allocated line builders
	for len(bc.lines) < bc.height {
		bc.lines = append(bc.lines, strings.Builder{})
	}

	// Reset line builders
	for i := 0; i < bc.height; i++ {
		bc.lines[i].Reset()
		bc.lines[i].Grow(bc.width * 4) // Estimate capacity for styled characters
	}

	// Calculate chart dimensions
	chartWidth := bc.width
	chartHeight := bc.height

	// Calculate the center line (split between upload and download)
	centerLine := chartHeight / 2

	// Calculate data points per character
	dataLen := len(bc.uploadData)
	downloadLen := len(bc.downloadData)
	if downloadLen > dataLen {
		dataLen = downloadLen
	}

	if dataLen == 0 {
		return bc.renderEmptyChart()
	}

	// Use different rendering approaches based on time scale
	if bc.timeScale == TimeScale1Min {
		// Original 1:1 rendering for 1-minute scale (no aggregation)
		for x := 0; x < chartWidth; x++ {
			// Calculate which data point this column represents (scrolling from right)
			dataIndex := dataLen - (chartWidth - x)

			// Get upload and download values for this column
			var upload, download uint64
			if dataIndex >= 0 && dataIndex < len(bc.uploadData) {
				upload = bc.uploadData[dataIndex]
			}
			if dataIndex >= 0 && dataIndex < len(bc.downloadData) {
				download = bc.downloadData[dataIndex]
			}

			// Render this column based on display mode
			if bc.overlayMode {
				bc.renderColumnOverlay(x, upload, download)
			} else {
				bc.renderColumn(x, upload, download, centerLine)
			}
		}
	} else {
		// Window-based aggregation for larger time scales
		bc.renderWithTimeWindows(chartWidth, centerLine)
	}

	// Combine all lines into final output
	for i := 0; i < bc.height; i++ {
		if i > 0 {
			bc.builder.WriteString("\n")
		}
		bc.builder.WriteString(bc.lines[i].String())
	}

	return bc.builder.String()
}

// renderEmptyChart renders an empty chart placeholder
func (bc *BrailleChart) renderEmptyChart() string {
	bc.builder.Reset()

	for y := 0; y < bc.height; y++ {
		if y > 0 {
			bc.builder.WriteString("\n")
		}
		// Empty space - no center line
		bc.builder.WriteString(strings.Repeat(" ", bc.width))
	}

	return bc.builder.String()
}

// renderWithTimeWindows renders the chart using fixed time windows for larger time scales
func (bc *BrailleChart) renderWithTimeWindows(chartWidth, centerLine int) {
	// Calculate window size (how many data points per column)
	timeScaleSeconds := bc.GetTimeScaleSeconds()
	windowSize := timeScaleSeconds / 60 // Each window represents this many data points
	if windowSize < 1 {
		windowSize = 1
	}

	dataLen := len(bc.uploadData)
	downloadLen := len(bc.downloadData)
	if downloadLen > dataLen {
		dataLen = downloadLen
	}

	if dataLen == 0 {
		// No data, render empty columns
		for x := 0; x < chartWidth; x++ {
			if bc.overlayMode {
				bc.renderColumnOverlay(x, 0, 0)
			} else {
				bc.renderColumn(x, 0, 0, centerLine)
			}
		}
		return
	}

	// The key insight: larger time scales should scroll EXACTLY like 1-minute scale
	// The rightmost column always shows the most recent data
	// Each column represents a window of aggregated data
	// Data scrolls from right to left as new data arrives

	for x := 0; x < chartWidth; x++ {
		// Calculate which data window this column represents
		// x=0 is leftmost, x=chartWidth-1 is rightmost (most recent)
		// Work backwards from the most recent data
		columnsFromRight := chartWidth - 1 - x
		
		// Calculate the data window for this column
		// Each column represents 'windowSize' data points
		windowEndIndex := dataLen - (columnsFromRight * windowSize)
		windowStartIndex := windowEndIndex - windowSize
		
		// Aggregate data within this window (use maximum value)
		var upload, download uint64

		// Find max upload in this window
		for i := windowStartIndex; i < windowEndIndex && i < len(bc.uploadData); i++ {
			if i >= 0 && bc.uploadData[i] > upload {
				upload = bc.uploadData[i]
			}
		}

		// Find max download in this window
		for i := windowStartIndex; i < windowEndIndex && i < len(bc.downloadData); i++ {
			if i >= 0 && bc.downloadData[i] > download {
				download = bc.downloadData[i]
			}
		}

		// Render this column based on display mode
		if bc.overlayMode {
			bc.renderColumnOverlay(x, upload, download)
		} else {
			bc.renderColumn(x, upload, download, centerLine)
		}
	}
}

// SetTimeScale sets the time scale directly (for debugging)
func (bc *BrailleChart) SetTimeScale(timeScale TimeScale) {
	bc.timeScale = timeScale
}
