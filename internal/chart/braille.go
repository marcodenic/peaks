// Package chart provides braille chart rendering functionality
//
// This package implements high-resolution braille chart rendering for terminal displays.
// It creates split-axis charts with upload data below and download data above a
// central axis, using Unicode braille characters for detailed visualization.
package chart

import (
	"fmt"
	"math"
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

	// Scaling constants
	logBase     = 10.0   // Base for logarithmic scaling
	minLogValue = 1024.0 // Minimum value for log scaling (1KB)
)

// ScalingMode defines how the chart scales data
type ScalingMode int

const (
	ScalingLinear ScalingMode = iota
	ScalingLogarithmic
	ScalingSquareRoot
)

// ColorGradient represents a color gradient configuration
type ColorGradient struct {
	Steps []lipgloss.Color
}

// Optimization: pre-calculated dot patterns as package constants
var dotPatterns = [4]int{
	0x01 | 0x08, // dots 0,3 (row 0)
	0x02 | 0x10, // dots 1,4 (row 1)
	0x04 | 0x20, // dots 2,5 (row 2)
	0x40 | 0x80, // dots 6,7 (row 3)
}

var (
	// Chart-specific styles for braille characters - base colors
	baseUploadColor   = lipgloss.Color("#F87171") // Red for upload
	baseDownloadColor = lipgloss.Color("#34D399") // Green for download

	// Color gradients for height-based shading (lighter as they get taller)
	uploadGradient = ColorGradient{
		Steps: []lipgloss.Color{
			lipgloss.Color("#7F1D1D"), // Dark red (bottom)
			lipgloss.Color("#B91C1C"), // Medium-dark red
			lipgloss.Color("#DC2626"), // Medium red
			lipgloss.Color("#EF4444"), // Medium-light red
			lipgloss.Color("#F87171"), // Light red
			lipgloss.Color("#FCA5A5"), // Very light red (top)
		},
	}

	downloadGradient = ColorGradient{
		Steps: []lipgloss.Color{
			lipgloss.Color("#064E3B"), // Dark green (bottom)
			lipgloss.Color("#047857"), // Medium-dark green
			lipgloss.Color("#059669"), // Medium green
			lipgloss.Color("#10B981"), // Medium-light green
			lipgloss.Color("#34D399"), // Light green
			lipgloss.Color("#6EE7B7"), // Very light green (top)
		},
	}

	// Yellow gradient for overlay overlap areas (dark to light from top to bottom)
	overlapGradient = ColorGradient{
		Steps: []lipgloss.Color{
			lipgloss.Color("#713F12"), // Very dark yellow/brown (darkest)
			lipgloss.Color("#92400E"), // Dark yellow
			lipgloss.Color("#B45309"), // Medium-dark yellow
			lipgloss.Color("#D97706"), // Medium yellow
			lipgloss.Color("#F59E0B"), // Medium-light yellow
			lipgloss.Color("#FBBF24"), // Light yellow
			lipgloss.Color("#FCD34D"), // Very light yellow
			lipgloss.Color("#FDE68A"), // Extremely light yellow (lightest)
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
		scalingMode: ScalingLogarithmic,                          // Default to logarithmic scaling
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

// SetScalingMode sets the scaling mode for the chart
func (bc *BrailleChart) SetScalingMode(mode ScalingMode) {
	bc.scalingMode = mode
}

// GetScalingMode returns the current scaling mode
func (bc *BrailleChart) GetScalingMode() ScalingMode {
	return bc.scalingMode
}

// CycleScalingMode cycles through available scaling modes
func (bc *BrailleChart) CycleScalingMode() ScalingMode {
	switch bc.scalingMode {
	case ScalingLinear:
		bc.scalingMode = ScalingLogarithmic
	case ScalingLogarithmic:
		bc.scalingMode = ScalingSquareRoot
	case ScalingSquareRoot:
		bc.scalingMode = ScalingLinear
	default:
		bc.scalingMode = ScalingLinear
	}
	return bc.scalingMode
}

// GetScalingModeName returns a human-readable name for the current scaling mode
func (bc *BrailleChart) GetScalingModeName() string {
	switch bc.scalingMode {
	case ScalingLinear:
		return "Linear"
	case ScalingLogarithmic:
		return "Logarithmic"
	case ScalingSquareRoot:
		return "Square Root"
	default:
		return "Unknown"
	}
}

// scaleValue applies the current scaling mode to a value
func (bc *BrailleChart) scaleValue(value uint64, maxValue uint64) float64 {
	if value == 0 {
		return 0
	}

	switch bc.scalingMode {
	case ScalingLinear:
		return float64(value) / float64(maxValue)

	case ScalingLogarithmic:
		// Ensure minimum value for log scaling
		val := math.Max(float64(value), minLogValue)
		maxVal := math.Max(float64(maxValue), minLogValue)

		// Apply logarithmic scaling
		logVal := math.Log10(val)
		logMax := math.Log10(maxVal)
		logMin := math.Log10(minLogValue)

		// Normalize to 0-1 range
		return (logVal - logMin) / (logMax - logMin)

	case ScalingSquareRoot:
		return math.Sqrt(float64(value)) / math.Sqrt(float64(maxValue))

	default:
		return float64(value) / float64(maxValue)
	}
}

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
	// Calculate heights for upload and download using new scaling
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
		bc.lines[y].WriteString(char)
	}
}

// renderColumnOverlay renders a single column in overlay mode
func (bc *BrailleChart) renderColumnOverlay(x int, upload, download uint64) {
	// Calculate heights for upload and download from bottom of chart using new scaling
	fullHeight := bc.height * brailleDots
	fullHeightFloat := float64(fullHeight)

	// Apply scaling to the values
	uploadScale := bc.scaleValue(upload, bc.maxValue)
	downloadScale := bc.scaleValue(download, bc.maxValue)

	uploadHeight := int(uploadScale * fullHeightFloat)
	downloadHeight := int(downloadScale * fullHeightFloat)

	// Clamp values
	if uploadHeight > fullHeight {
		uploadHeight = fullHeight
	}
	if downloadHeight > fullHeight {
		downloadHeight = fullHeight
	}

	// Render each row in this column
	for y := 0; y < bc.height; y++ {
		char := bc.createBrailleCharForOverlay(y, uploadHeight, downloadHeight, fullHeight, uploadScale, downloadScale)
		bc.lines[y].WriteString(char)
	}
}

// createBrailleCharForLineSplit creates a braille character for a specific line with split axis
func (bc *BrailleChart) createBrailleCharForLineSplit(line, uploadHeight, downloadHeight, halfHeight int, uploadScale, downloadScale float64) string {
	// Optimization: early return for empty characters
	if uploadHeight == 0 && downloadHeight == 0 {
		return " "
	}

	// Base braille character
	base := brailleBase
	var dots int

	hasUpload := false
	hasDownload := false
	var uploadGradientPos, downloadGradientPos float64

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
				// Calculate gradient position based on ABSOLUTE distance from axis for horizontal consistency
				// For download in split mode: 0.0 = light (0.0), 1.0 = dark (1.0)
				// distanceFromAxis ranges from 1 (just above axis) to downloadHeight (top of column)
				// We want: axis = light (0.0), away from axis = dark (1.0)
				downloadGradientPos = float64(distanceFromAxis-1) / float64(halfHeight-1)
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
				// Calculate gradient position based on ABSOLUTE distance from axis for horizontal consistency
				// For upload in split mode: 0.0 = light (0.0), 1.0 = dark (1.0)
				// distanceFromAxis ranges from 0 (at axis) to uploadHeight-1 (bottom of column)
				// We want: axis = light (0.0), away from axis = dark (1.0)
				uploadGradientPos = float64(distanceFromAxis) / float64(halfHeight-1)
			}
		}
	}

	// Optimization: skip character creation if no dots
	if dots == 0 {
		return " "
	}

	// Create the character
	char := rune(base + dots)

	// Use gradient styling based on vertical position within the column
	if hasUpload && hasDownload {
		// This shouldn't happen with split axis, but just in case - use upload style
		return bc.getStyledCharWithGradient(char, uploadGradientPos, true)
	} else if hasUpload {
		return bc.getStyledCharWithGradient(char, uploadGradientPos, true)
	} else if hasDownload {
		return bc.getStyledCharWithGradient(char, downloadGradientPos, false)
	}

	return string(char)
}

// createBrailleCharForOverlay creates a braille character for overlay mode
func (bc *BrailleChart) createBrailleCharForOverlay(line, uploadHeight, downloadHeight, fullHeight int, uploadScale, downloadScale float64) string {
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

	// Optimization: skip character creation if no dots
	if uploadDots == 0 && downloadDots == 0 {
		return " "
	}

	// Create the character with all dots
	char := rune(base + (uploadDots | downloadDots))

	// Calculate gradient position based on ABSOLUTE position in chart for horizontal consistency
	// This ensures all columns have the same gradient regardless of their individual heights
	// We want: bottom = light (0.0), top = dark (1.0)
	// Since line 0 is at the top and line (height-1) is at the bottom, we need to invert
	gradientPos := 1.0 - (float64(lineTop + brailleDots/2) / float64(fullHeight-1))

	// Clamp gradient position
	if gradientPos < 0 {
		gradientPos = 0
	}
	if gradientPos > 1 {
		gradientPos = 1
	}

	// Determine color based on overlap status but use same gradient position for all
	if overlapDots != 0 {
		// Overlap area - use yellow gradient
		return bc.getStyledCharWithOverlapGradient(char, gradientPos)
	} else if uploadDots != 0 && downloadDots != 0 {
		// Both present but no overlap at dot level - use yellow gradient
		return bc.getStyledCharWithOverlapGradient(char, gradientPos)
	} else if uploadDots != 0 {
		// Upload-only area - use red gradient
		return bc.getStyledCharWithGradient(char, gradientPos, true)
	} else if downloadDots != 0 {
		// Download-only area - use green gradient
		return bc.getStyledCharWithGradient(char, gradientPos, false)
	}

	return " "
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
