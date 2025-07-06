package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BrailleChart creates beautiful braille-based charts for terminal display
type BrailleChart struct {
	width      int
	maxPoints  int
	uploadData []uint64
	downloadData []uint64
	maxValue   uint64
	minHeight  int
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
		maxPoints:    maxPoints,
		uploadData:   make([]uint64, 0, maxPoints),
		downloadData: make([]uint64, 0, maxPoints),
		maxValue:     1024, // Start with 1KB minimum scale
		minHeight:    8,    // Minimum chart height in rows
	}
}

// SetWidth sets the chart width
func (bc *BrailleChart) SetWidth(width int) {
	bc.width = width
	if bc.width < 20 {
		bc.width = 20
	}
}

// AddDataPoint adds a new data point to the chart
func (bc *BrailleChart) AddDataPoint(upload, download uint64) {
	// Add new data points
	bc.uploadData = append(bc.uploadData, upload)
	bc.downloadData = append(bc.downloadData, download)
	
	// Remove old data points if we exceed max
	if len(bc.uploadData) > bc.maxPoints {
		bc.uploadData = bc.uploadData[1:]
	}
	if len(bc.downloadData) > bc.maxPoints {
		bc.downloadData = bc.downloadData[1:]
	}
	
	// Update max value for scaling
	bc.updateMaxValue()
}

// updateMaxValue updates the maximum value for scaling
func (bc *BrailleChart) updateMaxValue() {
	var currentMax uint64
	
	// Find the maximum value in recent data (more stable scaling)
	for _, val := range bc.uploadData {
		if val > currentMax {
			currentMax = val
		}
	}
	for _, val := range bc.downloadData {
		if val > currentMax {
			currentMax = val
		}
	}
	
	// Use more stable scaling that doesn't jump around
	if currentMax < 1024 {
		bc.maxValue = 1024
	} else if currentMax > bc.maxValue {
		// Only scale up, and do it gradually
		bc.maxValue = currentMax * 2
	} else if currentMax < bc.maxValue/4 && bc.maxValue > 1024 {
		// Only scale down if the current max is much smaller, and gradually
		bc.maxValue = bc.maxValue / 2
	}
	
	// Reasonable bounds
	if bc.maxValue > 100*1024*1024 { // 100MB/s max
		bc.maxValue = 100 * 1024 * 1024
	}
}

// Reset clears all data
func (bc *BrailleChart) Reset() {
	bc.uploadData = bc.uploadData[:0]
	bc.downloadData = bc.downloadData[:0]
	bc.maxValue = 1024
}

// Render creates the beautiful braille chart with split axis
func (bc *BrailleChart) Render() string {
	if len(bc.uploadData) == 0 && len(bc.downloadData) == 0 {
		return bc.renderEmptyChart()
	}
	
	// Calculate dimensions
	chartWidth := bc.width - 10 // Leave space for labels
	if chartWidth < 20 {
		chartWidth = 20
	}
	
	// Create the chart
	chart := bc.renderBrailleChart(chartWidth)
	
	// Add horizontal axis line in the middle
	axisStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Faint(true)
	axis := axisStyle.Render(strings.Repeat("━", chartWidth))
	
	// Split the chart at the middle to insert the axis
	chartLines := strings.Split(chart, "\n")
	if len(chartLines) >= 3 {
		// Insert axis in the middle (between line 2 and 3 for a 6-line chart)
		midPoint := len(chartLines) / 2
		var result []string
		result = append(result, chartLines[:midPoint]...)
		result = append(result, axis)
		result = append(result, chartLines[midPoint:]...)
		chart = strings.Join(result, "\n")
	}
	
	// Add scale information
	scale := bc.renderScale()
	
	// Add legend
	legend := bc.renderLegend()
	
	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Left,
		legend,
		chart,
		scale,
	)
}

// renderEmptyChart renders a placeholder when no data is available
func (bc *BrailleChart) renderEmptyChart() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)
	
	// Create a simple animated-looking placeholder
	width := bc.width
	if width < 20 {
		width = 40
	}
	
	var lines []string
	lines = append(lines, emptyStyle.Render("Collecting network data..."))
	
	// Add some placeholder braille characters
	placeholder := strings.Repeat("⠤", width/4) + strings.Repeat("⠤", width/4) + strings.Repeat("⠤", width/4) + strings.Repeat("⠤", width/4)
	lines = append(lines, strings.Repeat("⠀", width))
	lines = append(lines, strings.Repeat("⠀", width))
	lines = append(lines, placeholder[:width])
	lines = append(lines, strings.Repeat("⠀", width))
	lines = append(lines, strings.Repeat("⠀", width))
	lines = append(lines, emptyStyle.Render("Chart will appear here shortly"))
	
	return strings.Join(lines, "\n")
}

// renderBrailleChart creates the actual braille chart with split axis
func (bc *BrailleChart) renderBrailleChart(width int) string {
	if len(bc.uploadData) == 0 && len(bc.downloadData) == 0 {
		return bc.renderEmptyChart()
	}
	
	// Chart dimensions
	chartHeight := 6  // Number of text lines for the chart
	dotsPerLine := 4  // Braille has 4 vertical dots per character
	totalHeight := chartHeight * dotsPerLine // Total vertical resolution
	halfHeight := totalHeight / 2 // Split point for upload/download
	
	// Get the data to display (last 'width' points)
	dataLen := len(bc.uploadData)
	if len(bc.downloadData) > dataLen {
		dataLen = len(bc.downloadData)
	}
	
	startIdx := 0
	if dataLen > width {
		startIdx = dataLen - width
	}
	
	// Create the chart grid
	lines := make([]string, chartHeight)
	
	for col := 0; col < width; col++ {
		dataIdx := startIdx + col
		
		var uploadVal, downloadVal uint64
		if dataIdx < len(bc.uploadData) && dataIdx >= 0 {
			uploadVal = bc.uploadData[dataIdx]
		}
		if dataIdx < len(bc.downloadData) && dataIdx >= 0 {
			downloadVal = bc.downloadData[dataIdx]
		}
		
		// Convert to chart heights (0 to halfHeight for each direction)
		uploadHeight := int(float64(uploadVal) / float64(bc.maxValue) * float64(halfHeight))
		downloadHeight := int(float64(downloadVal) / float64(bc.maxValue) * float64(halfHeight))
		
		if uploadHeight > halfHeight {
			uploadHeight = halfHeight
		}
		if downloadHeight > halfHeight {
			downloadHeight = halfHeight
		}
		
		// Create braille character for each line
		for line := 0; line < chartHeight; line++ {
			char := bc.createBrailleCharForLineSplit(line, uploadHeight, downloadHeight, halfHeight, dotsPerLine)
			lines[line] += char
		}
	}
	
	// Return lines in their natural order (top to bottom)
	return strings.Join(lines, "\n")
}

// createBrailleCharForLineSplit creates a braille character for a specific line with split axis
func (bc *BrailleChart) createBrailleCharForLineSplit(line, uploadHeight, downloadHeight, halfHeight, dotsPerLine int) string {
	// Base braille character
	base := 0x2800
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
				// Fill both columns for area effect
				switch dotRow {
				case 0:
					dots |= 0x01 // dot 0 (left column)
					dots |= 0x08 // dot 3 (right column)
				case 1:
					dots |= 0x02 // dot 1 (left column)
					dots |= 0x10 // dot 4 (right column)
				case 2:
					dots |= 0x04 // dot 2 (left column)
					dots |= 0x20 // dot 5 (right column)
				case 3:
					dots |= 0x40 // dot 6 (left column)
					dots |= 0x80 // dot 7 (right column)
				}
			}
		}
		
		// Check if this dot should be filled for upload (below axis)
		// Upload fills from axis down to bottom
		if absoluteDotPos >= halfHeight {
			// We're below the axis - check if within upload area
			distanceFromAxis := absoluteDotPos - halfHeight
			if distanceFromAxis < uploadHeight {
				hasUpload = true
				// Fill both columns for area effect
				switch dotRow {
				case 0:
					dots |= 0x01 // dot 0 (left column)
					dots |= 0x08 // dot 3 (right column)
				case 1:
					dots |= 0x02 // dot 1 (left column)
					dots |= 0x10 // dot 4 (right column)
				case 2:
					dots |= 0x04 // dot 2 (left column)
					dots |= 0x20 // dot 5 (right column)
				case 3:
					dots |= 0x40 // dot 6 (left column)
					dots |= 0x80 // dot 7 (right column)
				}
			}
		}
	}
	
	// Create the character
	char := rune(base + dots)
	charStr := string(char)
	
	// Color it based on what data it represents (no overlap since they're in different areas)
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

// renderScale creates a scale indicator for the split-axis chart
func (bc *BrailleChart) renderScale() string {
	scaleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Faint(true)
	
	return scaleStyle.Render(
		"Scale: ±" + formatBandwidth(bc.maxValue) + " (download above axis, upload below)",
	)
}

// renderLegend creates a legend for the split-axis chart
func (bc *BrailleChart) renderLegend() string {
	legendStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D1D5DB")).
		Background(lipgloss.Color("#1F2937")).
		Padding(0, 1)
	
	upload := uploadStyle.Render("⠈ Upload (below)")
	download := downloadStyle.Render("⠈ Download (above)")
	axis := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Render("━ Axis")
	
	return legendStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			download, "  ", axis, "  ", upload,
		),
	)
}
