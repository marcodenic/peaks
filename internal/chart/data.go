// Package chart provides data management functionality for braille charts
package chart

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
