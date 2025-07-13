package main

import (
	"fmt"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Debug Window Distribution")
	fmt.Println("========================")

	bc := chart.NewBrailleChart(200)
	bc.SetWidth(20) // Use smaller width for easier debugging
	bc.SetHeight(6)

	// Add 50 data points
	fmt.Println("Adding 50 data points...")
	for i := 0; i < 50; i++ {
		upload := uint64(1000 + i*100)
		download := uint64(800 + i*80)
		bc.AddDataPoint(upload, download)
	}

	fmt.Printf("Chart width: %d\n", bc.GetWidth())
	fmt.Printf("Data length: %d\n", bc.GetDataLength())

	// Test 5-minute scale (window size = 5)
	bc.SetTimeScale(chart.TimeScale5Min)
	timeScaleSeconds := bc.GetTimeScaleSeconds()
	windowSize := timeScaleSeconds / 60
	
	fmt.Printf("Time scale: %s\n", bc.GetTimeScaleName())
	fmt.Printf("Time scale seconds: %d\n", timeScaleSeconds)
	fmt.Printf("Window size: %d\n", windowSize)

	chartWidth := bc.GetWidth()
	dataLen := bc.GetDataLength()

	fmt.Println("\nWindow calculation for each column:")
	for x := 0; x < chartWidth; x++ {
		// This is the same logic as in renderWithTimeWindows
		columnsFromRight := chartWidth - 1 - x
		windowEndOffset := columnsFromRight * windowSize
		windowEndIndex := dataLen - windowEndOffset
		windowStartIndex := windowEndIndex - windowSize
		
		fmt.Printf("Column %2d: columnsFromRight=%2d, windowEndOffset=%2d, window=[%2d:%2d]", 
			x, columnsFromRight, windowEndOffset, windowStartIndex, windowEndIndex)
		
		if windowStartIndex < 0 || windowEndIndex <= 0 {
			fmt.Printf(" -> NO DATA")
		} else if windowStartIndex >= dataLen {
			fmt.Printf(" -> OUT OF RANGE")
		} else {
			// Calculate what data this window would contain
			actualStart := windowStartIndex
			if actualStart < 0 {
				actualStart = 0
			}
			actualEnd := windowEndIndex
			if actualEnd > dataLen {
				actualEnd = dataLen
			}
			if actualEnd > actualStart {
				fmt.Printf(" -> data indices [%2d:%2d]", actualStart, actualEnd)
			} else {
				fmt.Printf(" -> EMPTY WINDOW")
			}
		}
		fmt.Println()
	}

	// Now render and show which columns have data
	rendered := bc.Render()
	lines := rendered
	if lines != "" {
		firstLine := ""
		for _, char := range lines {
			if char == '\n' {
				break
			}
			firstLine += string(char)
		}
		
		fmt.Printf("\nFirst line rendered (%d chars): '%s'\n", len(firstLine), firstLine)
		
		// Count non-space characters by position
		fmt.Println("Data distribution by column:")
		for i, char := range firstLine {
			if char != ' ' && char != '\x1b' { // Not space or escape
				fmt.Printf("  Column %d: '%c'\n", i, char)
			}
		}
	}
}
