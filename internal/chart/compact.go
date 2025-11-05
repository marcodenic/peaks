package chart

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderCompact renders a 2-line compact braille chart for terminal header use
// This is a simplified version that creates a horizontal visualization suitable
// for running at the top of a terminal while still allowing terminal use below.
func (bc *BrailleChart) RenderCompact(terminalWidth int) string {
	if len(bc.uploadData) == 0 && len(bc.downloadData) == 0 {
		return bc.renderEmptyCompact(terminalWidth)
	}

	// Update scaling based on currently visible data
	bc.updateMaxValue()

	// Compact mode is always 2 lines (8 braille rows, since each braille char = 4 vertical dots)
	const compactHeight = 2
	chartWidth := terminalWidth // Use exact full width

	if chartWidth < 10 {
		chartWidth = 10
	}

	// Prepare line builders - pre-allocate for performance
	line1 := strings.Builder{}
	line2 := strings.Builder{}
	line1.Grow(chartWidth * 20) // Account for ANSI color codes
	line2.Grow(chartWidth * 20)

	// Get data length
	dataLen := len(bc.uploadData)
	downloadLen := len(bc.downloadData)
	if downloadLen > dataLen {
		dataLen = downloadLen
	}

	if dataLen == 0 {
		return bc.renderEmptyCompact(terminalWidth)
	}

	// Define colors (same as full mode)
	uploadColor := lipgloss.Color("#EF4444")   // Red
	downloadColor := lipgloss.Color("#10B981") // Green
	overlapColor := lipgloss.Color("#EAB308")  // Yellow for overlap
	bgColor := lipgloss.Color("#374151")       // Grey background

	uploadStyle := lipgloss.NewStyle().Foreground(uploadColor)
	downloadStyle := lipgloss.NewStyle().Foreground(downloadColor)
	overlapStyle := lipgloss.NewStyle().Foreground(overlapColor)
	bgStyle := lipgloss.NewStyle().Foreground(bgColor)

	// Render each column (same logic as full chart)
	for x := 0; x < chartWidth; x++ {
		// Calculate which data point this column represents (scrolling from right)
		dataIndex := dataLen - (chartWidth - x)

		var uploadVal, downloadVal uint64
		
		// Get upload and download values for this column (use 0 if no data yet)
		if dataIndex >= 0 && dataIndex < len(bc.uploadData) {
			uploadVal = bc.uploadData[dataIndex]
		}
		if dataIndex >= 0 && dataIndex < len(bc.downloadData) {
			downloadVal = bc.downloadData[dataIndex]
		}

		// Scale values (returns 0-1 normalized values)
		uploadScaled := bc.scaleValue(uploadVal, bc.maxValue)
		downloadScaled := bc.scaleValue(downloadVal, bc.maxValue)

		// Calculate heights
		// In split mode: each line has 4 dots, so scale to 0-4
		// In overlay mode: 8 dots total (2 lines * 4 dots)
		if bc.overlayMode {
			maxHeight := compactHeight * 4 // 8 dots total
			uploadHeight := int(uploadScaled * float64(maxHeight))
			downloadHeight := int(downloadScaled * float64(maxHeight))

			if uploadHeight > maxHeight {
				uploadHeight = maxHeight
			}
			if downloadHeight > maxHeight {
				downloadHeight = maxHeight
			}

			// Render column based on mode
			// Overlay mode: both start from bottom
			col1Char, col2Char := bc.renderCompactColumnOverlay(uploadHeight, downloadHeight, maxHeight)
			
			// Determine color based on overlap
			var style1, style2 lipgloss.Style
			if uploadHeight > 0 && downloadHeight > 0 {
				// Check if there's overlap in each character
				if uploadHeight >= 4 && downloadHeight >= 4 {
					style1 = overlapStyle // Both reach into first char
				} else if uploadHeight > downloadHeight {
					style1 = uploadStyle
				} else {
					style1 = downloadStyle
				}

				if uploadHeight > 4 && downloadHeight > 4 {
					style2 = overlapStyle // Both reach into second char
				} else if uploadHeight > downloadHeight && uploadHeight > 4 {
					style2 = uploadStyle
				} else if downloadHeight > 4 {
					style2 = downloadStyle
				} else {
					style2 = bgStyle
				}
			} else if uploadHeight > 0 {
				style1 = uploadStyle
				style2 = uploadStyle
			} else if downloadHeight > 0 {
				style1 = downloadStyle
				style2 = downloadStyle
			} else {
				style1 = bgStyle
				style2 = bgStyle
			}

			line2.WriteString(style1.Render(string(col1Char)))
			line1.WriteString(style2.Render(string(col2Char)))
		} else {
			// Split mode: each line has 4 dots max
			maxHeightPerLine := 4
			uploadHeight := int(uploadScaled * float64(maxHeightPerLine))
			downloadHeight := int(downloadScaled * float64(maxHeightPerLine))

			if uploadHeight > maxHeightPerLine {
				uploadHeight = maxHeightPerLine
			}
			if downloadHeight > maxHeightPerLine {
				downloadHeight = maxHeightPerLine
			}

			// Split mode: download grows upward from middle, upload grows downward from middle
			// Line 1 (top row): shows download (GREEN) growing upward toward top edge
			// Line 2 (bottom row): shows upload (RED) growing downward toward bottom edge
			topChar, bottomChar := bc.renderCompactColumnSplit(uploadHeight, downloadHeight)
			
			// Color code each character
			var styleTop, styleBottom lipgloss.Style
			
			// Top row shows download (green) growing upward
			if downloadHeight > 0 {
				styleTop = downloadStyle
			} else {
				styleTop = bgStyle
			}
			
			// Bottom row shows upload (red) growing downward
			if uploadHeight > 0 {
				styleBottom = uploadStyle
			} else {
				styleBottom = bgStyle
			}

			line1.WriteString(styleTop.Render(string(topChar)))
			line2.WriteString(styleBottom.Render(string(bottomChar)))
		}
	}

	// Combine lines
	return line1.String() + "\n" + line2.String()
}

// renderCompactColumnOverlay renders a column in overlay mode (both from bottom)
func (bc *BrailleChart) renderCompactColumnOverlay(uploadHeight, downloadHeight, maxHeight int) (rune, rune) {
	// Take the maximum of the two for visualization
	height := uploadHeight
	if downloadHeight > height {
		height = downloadHeight
	}

	// Each braille character represents 4 vertical dots
	// We have 2 characters (8 dots total)
	bottomChar := bc.getBrailleChar(height, 0, 4)    // Bottom 4 dots
	topChar := bc.getBrailleChar(height, 4, 8)       // Top 4 dots

	return bottomChar, topChar
}

// renderCompactColumnSplit renders a column in split mode
// Download (green) in line 1: fills from BOTTOM of line upward (close to red)
// Upload (red) in line 2: fills from TOP of line downward (close to green)
func (bc *BrailleChart) renderCompactColumnSplit(uploadHeight, downloadHeight int) (rune, rune) {
	// Line 1 (top/green): use getBrailleChar which now fills from BOTTOM up
	topChar := bc.getBrailleChar(downloadHeight, 0, 4)
	
	// Line 2 (bottom/red): use getBrailleCharInverted which fills from TOP down
	bottomChar := bc.getBrailleCharInverted(uploadHeight, 0, 4)

	return topChar, bottomChar
}

// getBrailleChar returns a braille character for a given height within a range
// height: 0-8, startDot: starting position, endDot: ending position
// Fills from BOTTOM upward
func (bc *BrailleChart) getBrailleChar(height, startDot, endDot int) rune {
	if height <= startDot {
		return '⠀' // Empty braille
	}

	dotsInRange := height - startDot
	if dotsInRange > (endDot - startDot) {
		dotsInRange = endDot - startDot
	}

	// Braille patterns that fill from BOTTOM to TOP
	// Using BOTH columns (left + right) for fuller appearance
	// These patterns start at the BOTTOM and add dots upward
	patterns := []rune{
		'⠀', // 0 dots (empty)
		'⠤', // 1 row: dots 3,6 (bottom row)
		'⠶', // 2 rows: dots 2,3,5,6
		'⠿', // 3 rows: dots 1,2,3,4,5,6
		'⣿', // 4 rows: all 8 dots (full)
	}

	if dotsInRange >= len(patterns) {
		return patterns[len(patterns)-1]
	}
	return patterns[dotsInRange]
}

// getBrailleCharInverted returns a braille character that fills from TOP downward
func (bc *BrailleChart) getBrailleCharInverted(height, startDot, endDot int) rune {
	if height <= startDot {
		return '⠀' // Empty braille
	}

	dotsInRange := height - startDot
	if dotsInRange > (endDot - startDot) {
		dotsInRange = endDot - startDot
	}

	// Braille patterns that fill from TOP to BOTTOM
	// Using BOTH columns (left + right) for fuller appearance
	// These patterns start at the TOP and add dots downward
	patterns := []rune{
		'⠀', // 0 dots (empty)
		'⠉', // 1 row: dots 1,4 (top row)
		'⠛', // 2 rows: dots 1,2,4,5
		'⠿', // 3 rows: dots 1,2,3,4,5,6
		'⣿', // 4 rows: all 8 dots (full)
	}

	if dotsInRange >= len(patterns) {
		return patterns[len(patterns)-1]
	}
	return patterns[dotsInRange]
}

// renderEmptyCompact renders an empty compact chart
func (bc *BrailleChart) renderEmptyCompact(terminalWidth int) string {
	bgColor := lipgloss.Color("#374151")
	bgStyle := lipgloss.NewStyle().Foreground(bgColor)
	
	chartWidth := terminalWidth - 2
	if chartWidth < 10 {
		chartWidth = 10
	}

	emptyLine := bgStyle.Render(strings.Repeat("⠀", chartWidth))
	return emptyLine + "\n" + emptyLine
}
