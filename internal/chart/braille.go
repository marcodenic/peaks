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
	// Cached column data for stability
	columnCache map[int][]string // windowIndex -> rendered column lines
	lastCompleteWindow int       // last window index that was completed
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
		// Initialize caching for stability
		columnCache: make(map[int][]string),
		lastCompleteWindow: -1,
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

	// Calculate how many complete windows we have
	totalCompleteWindows := dataLen / windowSize
	hasPartialWindow := (dataLen % windowSize) != 0
	totalWindows := totalCompleteWindows
	if hasPartialWindow {
		totalWindows++
	}

	// Update cache for newly completed windows
	bc.updateColumnCache(totalCompleteWindows, windowSize, centerLine)

	// Calculate which windows to display
	firstVisibleWindow := 0
	if totalWindows > chartWidth {
		firstVisibleWindow = totalWindows - chartWidth
	}

	for x := 0; x < chartWidth; x++ {
		windowIndex := firstVisibleWindow + x
		
		// Check if this window is beyond our data
		if windowIndex >= totalWindows {
			// No data for this column
			if bc.overlayMode {
				bc.renderColumnOverlay(x, 0, 0)
			} else {
				bc.renderColumn(x, 0, 0, centerLine)
			}
			continue
		}

		// Use cached column if available (for completed windows)
		if cachedColumn, exists := bc.columnCache[windowIndex]; exists && windowIndex < totalCompleteWindows {
			// Use cached rendering for stability
			for y := 0; y < len(cachedColumn) && y < bc.height; y++ {
				bc.lines[y].WriteString(cachedColumn[y])
			}
			continue
		}

		// Calculate window boundaries for live rendering (incomplete windows only)
		windowStartIndex := windowIndex * windowSize
		windowEndIndex := windowStartIndex + windowSize
		if windowEndIndex > dataLen {
			windowEndIndex = dataLen
		}
		
		// Skip empty windows
		if windowStartIndex >= windowEndIndex {
			if bc.overlayMode {
				bc.renderColumnOverlay(x, 0, 0)
			} else {
				bc.renderColumn(x, 0, 0, centerLine)
			}
			continue
		}
		
		// Aggregate data within this window (live calculation for incomplete windows)
		var upload, download uint64

		// Find max upload in this window
		for i := windowStartIndex; i < windowEndIndex && i < len(bc.uploadData); i++ {
			if bc.uploadData[i] > upload {
				upload = bc.uploadData[i]
			}
		}

		// Find max download in this window
		for i := windowStartIndex; i < windowEndIndex && i < len(bc.downloadData); i++ {
			if bc.downloadData[i] > download {
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

// updateColumnCache updates the cache for newly completed windows
func (bc *BrailleChart) updateColumnCache(totalCompleteWindows, windowSize, centerLine int) {
	for windowIndex := bc.lastCompleteWindow + 1; windowIndex < totalCompleteWindows; windowIndex++ {
		// This window is now complete, cache its rendering
		windowStartIndex := windowIndex * windowSize
		windowEndIndex := windowStartIndex + windowSize
		
		// Aggregate data for this completed window
		var upload, download uint64
		
		// Find max upload in this window
		for i := windowStartIndex; i < windowEndIndex && i < len(bc.uploadData); i++ {
			if bc.uploadData[i] > upload {
				upload = bc.uploadData[i]
			}
		}

		// Find max download in this window
		for i := windowStartIndex; i < windowEndIndex && i < len(bc.downloadData); i++ {
			if bc.downloadData[i] > download {
				download = bc.downloadData[i]
			}
		}

		// Render this window to cache
		cachedColumn := bc.renderColumnToCache(upload, download, centerLine)
		bc.columnCache[windowIndex] = cachedColumn
	}
	
	bc.lastCompleteWindow = totalCompleteWindows - 1
}

// renderColumnToCache renders a column and returns the result as a slice of strings
func (bc *BrailleChart) renderColumnToCache(upload, download uint64, centerLine int) []string {
	// Create temporary builders for this column
	tempLines := make([]strings.Builder, bc.height)
	
	// Calculate heights for upload and download using scaling
	halfHeight := centerLine * brailleDots
	halfHeightFloat := float64(halfHeight)

	// Apply scaling to the values
	uploadScale := bc.scaleValue(upload, bc.maxValue)
	downloadScale := bc.scaleValue(download, bc.maxValue)

	uploadHeight := int(uploadScale * halfHeightFloat)
	downloadHeight := int(downloadScale * halfHeightFloat)

	// Clamp values
	if uploadHeight > halfHeight {
		uploadHeight = halfHeight
	}
	if downloadHeight > halfHeight {
		downloadHeight = halfHeight
	}

	// Render each row in this column
	for y := 0; y < bc.height; y++ {
		char := bc.createBrailleCharForLineSplit(y, uploadHeight, downloadHeight, halfHeight, uploadScale, downloadScale)
		tempLines[y].WriteString(char)
	}
	
	// Convert to strings
	result := make([]string, bc.height)
	for y := 0; y < bc.height; y++ {
		result[y] = tempLines[y].String()
	}
	
	return result
}

// SetTimeScale sets the time scale directly (for debugging)
func (bc *BrailleChart) SetTimeScale(timeScale TimeScale) {
	bc.timeScale = timeScale
}
