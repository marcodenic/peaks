package main

import (
	"fmt"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	// Create a chart with a small width for easier testing
	bc := chart.NewBrailleChart(20) // 20 columns wide
	bc.SetHeight(10) // 10 rows tall

	// Simulate adding 100 data points over time
	fmt.Println("Testing Time Scale Stability")
	fmt.Println("============================")

	// Add some initial data
	for i := 0; i < 50; i++ {
		upload := uint64(100 + i*10)
		download := uint64(80 + i*8)
		bc.AddDataPoint(upload, download)
	}

	// Test each time scale
	timeScales := []chart.TimeScale{
		chart.TimeScale1Min,
		chart.TimeScale3Min,
		chart.TimeScale5Min,
		chart.TimeScale10Min,
		chart.TimeScale15Min,
		chart.TimeScale30Min,
		chart.TimeScale60Min,
	}

	timeScaleNames := []string{
		"1 minute",
		"3 minutes", 
		"5 minutes",
		"10 minutes",
		"15 minutes",
		"30 minutes", 
		"60 minutes",
	}

	for scaleIdx, scale := range timeScales {
		fmt.Printf("\n--- Testing %s time scale ---\n", timeScaleNames[scaleIdx])
		
		bc.SetTimeScale(scale)
		
		// Capture the current rendering state
		beforeData := captureColumnData(bc)
		
		// Add one new data point
		bc.AddDataPoint(uint64(200), uint64(180))
		
		// Capture the state after adding data
		afterData := captureColumnData(bc)
		
		// Analyze the differences
		unchangedColumns := 0
		changedColumns := 0
		
		for col := 0; col < len(beforeData) && col < len(afterData); col++ {
			if beforeData[col].upload == afterData[col].upload && 
			   beforeData[col].download == afterData[col].download {
				unchangedColumns++
			} else {
				changedColumns++
				if col < len(beforeData)-1 { // Not the rightmost column
					fmt.Printf("  WARNING: Column %d changed (not rightmost): before=(%d,%d) after=(%d,%d)\n", 
						col, beforeData[col].upload, beforeData[col].download,
						afterData[col].upload, afterData[col].download)
				}
			}
		}
		
		fmt.Printf("  Unchanged columns: %d\n", unchangedColumns)
		fmt.Printf("  Changed columns: %d\n", changedColumns)
		
		if scale == chart.TimeScale1Min {
			// For 1-minute scale, we expect exactly 1 column to change (scroll effect)
			if changedColumns == 1 {
				fmt.Printf("  ✓ PASS: Only rightmost column changed (expected for 1min scale)\n")
			} else {
				fmt.Printf("  ✗ FAIL: Expected 1 column to change, got %d\n", changedColumns)
			}
		} else {
			// For larger time scales, we expect at most 1 column to change (rightmost window update)
			if changedColumns <= 1 {
				fmt.Printf("  ✓ PASS: At most 1 column changed (stable windows)\n")
			} else {
				fmt.Printf("  ✗ FAIL: Too many columns changed (%d), bars will jump\n", changedColumns)
			}
		}
	}

	fmt.Println("\n--- Final Render Test ---")
	bc.SetTimeScale(chart.TimeScale5Min)
	rendered := bc.Render()
	fmt.Printf("Sample 5-minute chart rendering:\n%s\n", rendered)
}

type ColumnData struct {
	upload   uint64
	download uint64
}

// captureColumnData extracts the data values that each column represents
func captureColumnData(bc *chart.BrailleChart) []ColumnData {
	// We need to simulate the same logic used in the Render method
	// to see what data each column represents
	
	width := bc.GetWidth()
	dataLen := bc.GetDataLength()
	
	var columns []ColumnData
	
	if bc.GetTimeScale() == chart.TimeScale1Min {
		// Original 1:1 logic
		for x := 0; x < width; x++ {
			dataIndex := dataLen - (width - x)
			
			var upload, download uint64
			if dataIndex >= 0 {
				// We can't access the actual data arrays from outside the package,
				// so we'll use a simplified approach for this test
				upload = uint64(dataIndex * 10)   // Simulate
				download = uint64(dataIndex * 8)   // Simulate
			}
			
			columns = append(columns, ColumnData{upload, download})
		}
	} else {
		// Window-based logic
		timeScaleSeconds := bc.GetTimeScaleSeconds()
		windowSize := timeScaleSeconds / 60
		if windowSize < 1 {
			windowSize = 1
		}
		
		totalWindows := width
		totalDataPointsNeeded := totalWindows * windowSize
		startDataIndex := dataLen - totalDataPointsNeeded
		if startDataIndex < 0 {
			startDataIndex = 0
		}
		
		for x := 0; x < width; x++ {
			windowStartIndex := startDataIndex + (x * windowSize)
			windowEndIndex := windowStartIndex + windowSize
			
			// Simulate finding max in window
			var upload, download uint64
			for i := windowStartIndex; i < windowEndIndex; i++ {
				if i >= 0 && i < dataLen {
					val := uint64(i * 10)
					if val > upload {
						upload = val
					}
					val = uint64(i * 8)
					if val > download {
						download = val
					}
				}
			}
			
			columns = append(columns, ColumnData{upload, download})
		}
	}
	
	return columns
}

func getTimeScaleSeconds(scale chart.TimeScale) int {
	switch scale {
	case chart.TimeScale1Min:
		return 60
	case chart.TimeScale3Min:
		return 180
	case chart.TimeScale5Min:
		return 300
	case chart.TimeScale10Min:
		return 600
	case chart.TimeScale15Min:
		return 900
	case chart.TimeScale30Min:
		return 1800
	case chart.TimeScale60Min:
		return 3600
	default:
		return 60
	}
}
