// Package chart provides rendering functionality for braille charts
package chart

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
