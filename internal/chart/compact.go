package chart

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderCompact renders a 2-line compact braille chart for terminal header use
// This is a simplified version that creates a horizontal visualization suitable
// for running at the top of a terminal while still allowing terminal use below.
func (bc *BrailleChart) RenderCompact(terminalWidth int) string {
	return bc.RenderCompactWithSize(terminalWidth, 2)
}

// RenderCompactWithSize renders a compact braille chart with custom height
func (bc *BrailleChart) RenderCompactWithSize(terminalWidth int, compactHeight int) string {
	if len(bc.uploadData) == 0 && len(bc.downloadData) == 0 {
		return bc.renderEmptyCompact(terminalWidth, compactHeight)
	}

	// Update scaling based on currently visible data
	bc.updateMaxValue()

	chartWidth := terminalWidth // Use full terminal width

	if chartWidth < 10 {
		chartWidth = 10
	}

	// Prepare line builders based on height
	lines := make([]strings.Builder, compactHeight)
	for i := 0; i < compactHeight; i++ {
		lines[i].Grow(chartWidth * 20) // Account for ANSI color codes
	}

	// Get data length
	dataLen := len(bc.uploadData)
	downloadLen := len(bc.downloadData)
	if downloadLen > dataLen {
		dataLen = downloadLen
	}

	if dataLen == 0 {
		return bc.renderEmptyCompact(terminalWidth, compactHeight)
	}

	// Define colors (same as full mode)
	uploadColor := lipgloss.Color("#EF4444")   // Red
	downloadColor := lipgloss.Color("#10B981") // Green
	overlapColor := lipgloss.Color("#EAB308")  // Yellow for overlap
	bgColor := lipgloss.Color("#374151")       // Grey background

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
		// In split mode: top half shows download, bottom half shows upload
		// In overlay mode: all lines show both colors from bottom
		if bc.overlayMode {
			maxHeight := compactHeight * 4 // 4 dots per line
			uploadHeight := int(uploadScaled * float64(maxHeight))
			downloadHeight := int(downloadScaled * float64(maxHeight))

			if uploadHeight > maxHeight {
				uploadHeight = maxHeight
			}
			if downloadHeight > maxHeight {
				downloadHeight = maxHeight
			}

			// Render column for overlay mode - all lines from bottom
			bc.renderCompactColumnOverlayMultiLine(x, uploadHeight, downloadHeight, maxHeight, compactHeight, lines, uploadColor, downloadColor, overlapColor, bgColor)
		} else {
			// Split mode: top half (compactHeight/2) for download, bottom half for upload
			halfLines := compactHeight / 2
			maxHeightPerHalf := halfLines * 4 // 4 dots per line
			
			uploadHeight := int(uploadScaled * float64(maxHeightPerHalf))
			downloadHeight := int(downloadScaled * float64(maxHeightPerHalf))

			if uploadHeight > maxHeightPerHalf {
				uploadHeight = maxHeightPerHalf
			}
			if downloadHeight > maxHeightPerHalf {
				downloadHeight = maxHeightPerHalf
			}

			// Split mode: download in top half, upload in bottom half
			bc.renderCompactColumnSplitMultiLine(x, uploadHeight, downloadHeight, halfLines, lines, uploadColor, downloadColor, bgColor)
		}
	}

	// Combine lines
	result := strings.Builder{}
	for i := 0; i < len(lines); i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(lines[i].String())
	}
	return result.String()
}

// renderCompactColumnOverlayMultiLine renders a column in overlay mode with multiple lines
func (bc *BrailleChart) renderCompactColumnOverlayMultiLine(x, uploadHeight, downloadHeight, maxHeight, compactHeight int, lines []strings.Builder, uploadColor, downloadColor, overlapColor, bgColor lipgloss.Color) {
	uploadStyle := lipgloss.NewStyle().Foreground(uploadColor)
	downloadStyle := lipgloss.NewStyle().Foreground(downloadColor)
	overlapStyle := lipgloss.NewStyle().Foreground(overlapColor)
	bgStyle := lipgloss.NewStyle().Foreground(bgColor)
	
	// Render from bottom to top (line index compactHeight-1 is bottom)
	for lineIdx := 0; lineIdx < compactHeight; lineIdx++ {
		// Calculate which dots this line represents (from bottom)
		// Bottom line (lineIdx = compactHeight-1) has dots 0-3
		// Next line up has dots 4-7, etc.
		lineFromBottom := compactHeight - 1 - lineIdx
		dotStart := lineFromBottom * 4
		
		// Check if upload/download reach this line
		uploadInLine := uploadHeight > dotStart
		downloadInLine := downloadHeight > dotStart
		
		if !uploadInLine && !downloadInLine {
			// Empty line
			lines[lineIdx].WriteString(bgStyle.Render("⠀"))
			continue
		}
		
		// Calculate how many dots to fill in this line (0-4)
		uploadDotsInLine := 0
		downloadDotsInLine := 0
		
		if uploadInLine {
			uploadDotsInLine = uploadHeight - dotStart
			if uploadDotsInLine > 4 {
				uploadDotsInLine = 4
			}
		}
		
		if downloadInLine {
			downloadDotsInLine = downloadHeight - dotStart
			if downloadDotsInLine > 4 {
				downloadDotsInLine = 4
			}
		}
		
		// Use the max for the character
		dotsToFill := uploadDotsInLine
		if downloadDotsInLine > dotsToFill {
			dotsToFill = downloadDotsInLine
		}
		
		char := bc.getBrailleChar(dotsToFill, 0, 4)
		
		// Determine color based on overlap
		var style lipgloss.Style
		if uploadInLine && downloadInLine {
			style = overlapStyle // Both present = yellow
		} else if uploadInLine {
			style = uploadStyle // Upload only = red
		} else {
			style = downloadStyle // Download only = green
		}
		
		lines[lineIdx].WriteString(style.Render(string(char)))
	}
}

// renderCompactColumnSplitMultiLine renders a column in split mode with multiple lines
func (bc *BrailleChart) renderCompactColumnSplitMultiLine(x, uploadHeight, downloadHeight, halfLines int, lines []strings.Builder, uploadColor, downloadColor, bgColor lipgloss.Color) {
	uploadStyle := lipgloss.NewStyle().Foreground(uploadColor)
	downloadStyle := lipgloss.NewStyle().Foreground(downloadColor)
	bgStyle := lipgloss.NewStyle().Foreground(bgColor)
	
	totalLines := halfLines * 2
	
	// Top half: download (green) - grows UPWARD from center (line halfLines-1) toward top (line 0)
	for lineIdx := 0; lineIdx < halfLines; lineIdx++ {
		// Line halfLines-1 is at the center (axis), line 0 is at the top
		// Calculate distance from center axis
		distanceFromCenter := (halfLines - 1) - lineIdx
		dotStart := distanceFromCenter * 4
		
		// Check if download reaches this line (growing away from center)
		if downloadHeight > dotStart {
			// Calculate how many dots to fill (0-4)
			dotsInLine := downloadHeight - dotStart
			if dotsInLine > 4 {
				dotsInLine = 4
			}
			
			// Use normal braille (fills from bottom up) since we're growing upward from center
			char := bc.getBrailleChar(dotsInLine, 0, 4)
			lines[lineIdx].WriteString(downloadStyle.Render(string(char)))
		} else {
			lines[lineIdx].WriteString(bgStyle.Render("⠀"))
		}
	}
	
	// Bottom half: upload (red) - grows DOWNWARD from center (line halfLines) toward bottom (line totalLines-1)
	for lineIdx := halfLines; lineIdx < totalLines; lineIdx++ {
		// Line halfLines is at the center (axis), line totalLines-1 is at the bottom
		// Calculate distance from center axis
		distanceFromCenter := lineIdx - halfLines
		dotStart := distanceFromCenter * 4
		
		// Check if upload reaches this line (growing away from center)
		if uploadHeight > dotStart {
			// Calculate how many dots to fill (0-4)
			dotsInLine := uploadHeight - dotStart
			if dotsInLine > 4 {
				dotsInLine = 4
			}
			
			// Use inverted braille (fills from top down) since we're growing downward from center
			char := bc.getBrailleCharInverted(dotsInLine, 0, 4)
			lines[lineIdx].WriteString(uploadStyle.Render(string(char)))
		} else {
			lines[lineIdx].WriteString(bgStyle.Render("⠀"))
		}
	}
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
	// Dots 7,8 are the bottom row, then 3,6, then 2,5, then 1,4 at top
	patterns := []rune{
		'⠀', // 0 dots (empty)
		'⣀', // 1 row: dots 7,8 (bottom row)
		'⣤', // 2 rows: dots 3,6,7,8
		'⣶', // 3 rows: dots 2,3,5,6,7,8
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
func (bc *BrailleChart) renderEmptyCompact(terminalWidth int, compactHeight int) string {
	bgColor := lipgloss.Color("#374151")
	bgStyle := lipgloss.NewStyle().Foreground(bgColor)
	
	chartWidth := terminalWidth // Use full width
	if chartWidth < 10 {
		chartWidth = 10
	}

	emptyLine := bgStyle.Render(strings.Repeat("⠀", chartWidth))
	lines := make([]string, compactHeight)
	for i := 0; i < compactHeight; i++ {
		lines[i] = emptyLine
	}
	return strings.Join(lines, "\n")
}
