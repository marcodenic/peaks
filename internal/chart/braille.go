// Package chart provides braille chart rendering functionality
//
// This package implements high-resolution braille chart rendering for terminal displays.
// It creates split-axis charts with upload data below and download data above a
// central axis, using Unicode braille characters for detailed visualization.
package chart

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// Chart configuration constants
	MinChartHeight = 8                 // Minimum chart height in rows
	brailleDots    = 4                 // Braille has 4 vertical dots per character
	brailleBase    = 0x2800            // Base braille character code
	maxScaleLimit  = 100 * 1024 * 1024 // 100MB/s maximum scale

	// Optimization: pre-calculated constants
	maxBrailleChars    = 256 // Maximum number of braille characters (0x2800-0x28FF)
	defaultChartWidth  = 80
	defaultChartHeight = 20
	defaultMaxPoints   = 50
)

// Optimization: pre-calculated dot patterns as package constants
var dotPatterns = [4]int{
	0x01 | 0x08, // dots 0,3 (row 0)
	0x02 | 0x10, // dots 1,4 (row 1)
	0x04 | 0x20, // dots 2,5 (row 2)
	0x40 | 0x80, // dots 6,7 (row 3)
}

var (
	// Chart-specific styles for braille characters
	uploadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")). // Red for upload
			Bold(true)

	downloadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34D399")). // Green for download
			Bold(true)

	// Overlap style for overlay mode
	overlapStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FCD34D")). // Yellow for overlap
			Bold(true)

	// Optimization: character cache for styled braille characters
	uploadCharCache   = make(map[rune]string, 256)
	downloadCharCache = make(map[rune]string, 256)
	overlapCharCache  = make(map[rune]string, 256)
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
}

// DataPoint represents a single measurement point
type DataPoint struct {
	Upload   uint64
	Download uint64
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

// SetMaxPoints updates the maximum number of data points to maintain
// If reducing the limit, excess data is trimmed from the beginning
func (bc *BrailleChart) SetMaxPoints(maxPoints int) {
	if maxPoints <= 0 {
		return
	}

	oldMaxPoints := bc.maxPoints
	bc.maxPoints = maxPoints

	// If reducing the limit, trim excess data
	if maxPoints < oldMaxPoints {
		// Trim upload data if necessary
		if len(bc.uploadData) > maxPoints {
			bc.uploadData = bc.uploadData[len(bc.uploadData)-maxPoints:]
		}
		// Trim download data if necessary
		if len(bc.downloadData) > maxPoints {
			bc.downloadData = bc.downloadData[len(bc.downloadData)-maxPoints:]
		}
		// Recalculate max value after trimming
		bc.recalculateMax()
	}

	// Update the capacity of the pre-allocated slices if needed
	if maxPoints > cap(bc.uploadData) {
		newUploadData := make([]uint64, len(bc.uploadData), maxPoints)
		copy(newUploadData, bc.uploadData)
		bc.uploadData = newUploadData

		newDownloadData := make([]uint64, len(bc.downloadData), maxPoints)
		copy(newDownloadData, bc.downloadData)
		bc.downloadData = newDownloadData
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

// AddDataPoint adds a new data point to the chart
func (bc *BrailleChart) AddDataPoint(upload, download uint64) {
	// Update current max efficiently
	bc.updateCurrentMax(upload, download)

	// Add new data points
	bc.uploadData = append(bc.uploadData, upload)
	bc.downloadData = append(bc.downloadData, download)

	// Manage data size
	bc.trimDataIfNeeded()

	// Update scaling
	bc.updateMaxValue()
}

// updateCurrentMax efficiently tracks the current maximum value
func (bc *BrailleChart) updateCurrentMax(upload, download uint64) {
	if upload > bc.currentMax {
		bc.currentMax = upload
	}
	if download > bc.currentMax {
		bc.currentMax = download
	}
}

// trimDataIfNeeded removes old data points when exceeding capacity
func (bc *BrailleChart) trimDataIfNeeded() {
	if len(bc.uploadData) > bc.maxPoints {
		removedUpload := bc.uploadData[0]
		bc.uploadData = bc.uploadData[1:]

		// If we removed the max value, recalculate
		if removedUpload == bc.currentMax {
			bc.recalculateMax()
		}
	}

	if len(bc.downloadData) > bc.maxPoints {
		removedDownload := bc.downloadData[0]
		bc.downloadData = bc.downloadData[1:]

		// If we removed the max value, recalculate
		if removedDownload == bc.currentMax {
			bc.recalculateMax()
		}
	}
}

// recalculateMax recalculates the maximum value after removing data
func (bc *BrailleChart) recalculateMax() {
	bc.currentMax = 0
	for _, value := range bc.uploadData {
		if value > bc.currentMax {
			bc.currentMax = value
		}
	}
	for _, value := range bc.downloadData {
		if value > bc.currentMax {
			bc.currentMax = value
		}
	}
}

// updateMaxValue updates the chart's maximum value for scaling based on visible data
func (bc *BrailleChart) updateMaxValue() {
	// Calculate max value from only the currently visible data points
	visibleMax := bc.getVisibleDataMax()

	// Use exact maximum value with no headroom for full height utilization
	if visibleMax > 0 {
		bc.maxValue = visibleMax
	} else {
		bc.maxValue = 1024 // Minimum scale
	}

	// Ensure minimum scale
	if bc.maxValue < 1024 {
		bc.maxValue = 1024
	}

	// Apply reasonable upper bound
	if bc.maxValue > maxScaleLimit {
		bc.maxValue = maxScaleLimit
	}
}

// getVisibleDataMax calculates the maximum value from currently visible data points
func (bc *BrailleChart) getVisibleDataMax() uint64 {
	var maxVal uint64

	// Calculate which data points are currently visible
	dataLen := len(bc.uploadData)
	if downloadLen := len(bc.downloadData); downloadLen > dataLen {
		dataLen = downloadLen
	}

	if dataLen == 0 {
		return 0
	}

	// Check only the visible data points (rightmost chartWidth points)
	startIndex := 0
	if dataLen > bc.width {
		startIndex = dataLen - bc.width
	}

	// Find max in visible upload data
	for i := startIndex; i < len(bc.uploadData); i++ {
		if bc.uploadData[i] > maxVal {
			maxVal = bc.uploadData[i]
		}
	}

	// Find max in visible download data
	for i := startIndex; i < len(bc.downloadData); i++ {
		if bc.downloadData[i] > maxVal {
			maxVal = bc.downloadData[i]
		}
	}

	return maxVal
}

// Reset clears all data points and resets the chart
func (bc *BrailleChart) Reset() {
	bc.uploadData = bc.uploadData[:0]
	bc.downloadData = bc.downloadData[:0]
	bc.maxValue = 1024
	bc.currentMax = 0
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
	if downloadLen := len(bc.downloadData); downloadLen > dataLen {
		dataLen = downloadLen
	}

	if dataLen == 0 {
		return bc.renderEmptyChart()
	}

	// Render each column
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

	// Combine all lines into final output
	for i := 0; i < bc.height; i++ {
		if i > 0 {
			bc.builder.WriteString("\n")
		}
		bc.builder.WriteString(bc.lines[i].String())
	}

	return bc.builder.String()
}

// renderColumn renders a single column of the chart
func (bc *BrailleChart) renderColumn(x int, upload, download uint64, centerLine int) {
	// Calculate heights for upload and download using original logic
	halfHeight := centerLine * brailleDots
	maxValueFloat := float64(bc.maxValue)
	halfHeightFloat := float64(halfHeight)

	uploadHeight := int(float64(upload) / maxValueFloat * halfHeightFloat)
	downloadHeight := int(float64(download) / maxValueFloat * halfHeightFloat)

	// Clamp values
	if uploadHeight > halfHeight {
		uploadHeight = halfHeight
	}
	if downloadHeight > halfHeight {
		downloadHeight = halfHeight
	}

	// Render each row in this column
	for y := 0; y < bc.height; y++ {
		char := bc.createBrailleCharForLineSplit(y, uploadHeight, downloadHeight, halfHeight)
		bc.lines[y].WriteString(char)
	}
}

// renderColumnOverlay renders a single column in overlay mode
func (bc *BrailleChart) renderColumnOverlay(x int, upload, download uint64) {
	// Calculate heights for upload and download from bottom of chart
	fullHeight := bc.height * brailleDots
	maxValueFloat := float64(bc.maxValue)
	fullHeightFloat := float64(fullHeight)

	uploadHeight := int(float64(upload) / maxValueFloat * fullHeightFloat)
	downloadHeight := int(float64(download) / maxValueFloat * fullHeightFloat)

	// Clamp values
	if uploadHeight > fullHeight {
		uploadHeight = fullHeight
	}
	if downloadHeight > fullHeight {
		downloadHeight = fullHeight
	}

	// Render each row in this column
	for y := 0; y < bc.height; y++ {
		char := bc.createBrailleCharForOverlay(y, uploadHeight, downloadHeight, fullHeight)
		bc.lines[y].WriteString(char)
	}
}

// createBrailleCharForLineSplit creates a braille character for a specific line with split axis
func (bc *BrailleChart) createBrailleCharForLineSplit(line, uploadHeight, downloadHeight, halfHeight int) string {
	// Optimization: early return for empty characters
	if uploadHeight == 0 && downloadHeight == 0 {
		return " "
	}

	// Base braille character
	base := brailleBase
	var dots int

	hasUpload := false
	hasDownload := false

	// Calculate the vertical range of this braille character
	// Line 0 is at the top, line 5 is at the bottom (natural order)
	lineTop := line * brailleDots

	// Check each dot position in this braille character (4 dots vertically)
	for dotRow := 0; dotRow < brailleDots; dotRow++ {
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

	// Optimization: use cached styled characters
	if hasUpload && hasDownload {
		// This shouldn't happen with split axis, but just in case
		return bc.getStyledChar(char, true) // Default to upload style
	} else if hasUpload {
		return bc.getStyledChar(char, true)
	} else if hasDownload {
		return bc.getStyledChar(char, false)
	}

	return string(char)
}

// createBrailleCharForOverlay creates a braille character for overlay mode
func (bc *BrailleChart) createBrailleCharForOverlay(line, uploadHeight, downloadHeight, fullHeight int) string {
	// Optimization: early return for empty characters
	if uploadHeight == 0 && downloadHeight == 0 {
		return " "
	}

	// Base braille character
	base := brailleBase
	var uploadDots, downloadDots int

	// Calculate the vertical range of this braille character
	// Line 0 is at the top, but we fill from bottom
	lineTop := line * brailleDots

	// Check each dot position in this braille character (4 dots vertically)
	for dotRow := 0; dotRow < brailleDots; dotRow++ {
		// Calculate the absolute dot position in the chart (from top)
		absoluteDotPos := lineTop + dotRow

		// Convert to distance from bottom
		distanceFromBottom := fullHeight - absoluteDotPos

		// Check if this dot should be filled for upload
		if distanceFromBottom <= uploadHeight {
			uploadDots |= dotPatterns[dotRow]
		}

		// Check if this dot should be filled for download
		if distanceFromBottom <= downloadHeight {
			downloadDots |= dotPatterns[dotRow]
		}
	}

	// Determine final dots and styling
	overlapDots := uploadDots & downloadDots
	uploadOnlyDots := uploadDots & ^downloadDots
	downloadOnlyDots := downloadDots & ^uploadDots

	// Optimization: skip character creation if no dots
	if uploadDots == 0 && downloadDots == 0 {
		return " "
	}

	// Create characters for each type
	var result strings.Builder

	// If there's overlap, render the overlap in yellow
	if overlapDots != 0 {
		overlapChar := rune(base + overlapDots)
		result.WriteString(bc.getStyledCharOverlay(overlapChar, "overlap"))
	}

	// Render upload-only dots in red
	if uploadOnlyDots != 0 {
		uploadChar := rune(base + uploadOnlyDots)
		result.WriteString(bc.getStyledCharOverlay(uploadChar, "upload"))
	}

	// Render download-only dots in green
	if downloadOnlyDots != 0 {
		downloadChar := rune(base + downloadOnlyDots)
		result.WriteString(bc.getStyledCharOverlay(downloadChar, "download"))
	}

	// For overlapping positions, we need to combine the characters
	// Since we can't actually overlay characters, we'll prioritize overlap color
	if overlapDots != 0 {
		char := rune(base + (uploadDots | downloadDots))
		return bc.getStyledCharOverlay(char, "overlap")
	} else if uploadDots != 0 && downloadDots != 0 {
		// Both present but no overlap at dot level - shouldn't happen in overlay mode
		char := rune(base + (uploadDots | downloadDots))
		return bc.getStyledCharOverlay(char, "overlap")
	} else if uploadDots != 0 {
		char := rune(base + uploadDots)
		return bc.getStyledCharOverlay(char, "upload")
	} else if downloadDots != 0 {
		char := rune(base + downloadDots)
		return bc.getStyledCharOverlay(char, "download")
	}

	return " "
}

// getStyledChar returns a cached styled character or creates and caches it
func (bc *BrailleChart) getStyledChar(char rune, isUpload bool) string {
	if isUpload {
		if cached, exists := uploadCharCache[char]; exists {
			return cached
		}
		styled := uploadStyle.Render(string(char))
		uploadCharCache[char] = styled
		return styled
	} else {
		if cached, exists := downloadCharCache[char]; exists {
			return cached
		}
		styled := downloadStyle.Render(string(char))
		downloadCharCache[char] = styled
		return styled
	}
}

// getStyledCharOverlay returns a cached styled character for overlay mode
func (bc *BrailleChart) getStyledCharOverlay(char rune, mode string) string {
	switch mode {
	case "upload":
		if cached, exists := uploadCharCache[char]; exists {
			return cached
		}
		styled := uploadStyle.Render(string(char))
		uploadCharCache[char] = styled
		return styled
	case "download":
		if cached, exists := downloadCharCache[char]; exists {
			return cached
		}
		styled := downloadStyle.Render(string(char))
		downloadCharCache[char] = styled
		return styled
	case "overlap":
		if cached, exists := overlapCharCache[char]; exists {
			return cached
		}
		styled := overlapStyle.Render(string(char))
		overlapCharCache[char] = styled
		return styled
	default:
		return string(char)
	}
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

// GetMaxValue returns the current maximum value for scaling
func (bc *BrailleChart) GetMaxValue() uint64 {
	return bc.maxValue
}

// GetDataLength returns the number of data points currently stored
func (bc *BrailleChart) GetDataLength() int {
	uploadLen := len(bc.uploadData)
	downloadLen := len(bc.downloadData)
	if uploadLen > downloadLen {
		return uploadLen
	}
	return downloadLen
}
