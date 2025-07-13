package main

import (
	"fmt"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Window Debug Analysis")
	fmt.Println("=====================")
	
	c := chart.NewBrailleChart(1000)
	c.SetWidth(10) // Small width for easier analysis
	c.SetHeight(6)
	c.SetTimeScale(chart.TimeScale3Min) // 3-minute window = windowSize=3
	
	// Add data points one by one and show the window breakdown
	for i := 0; i < 10; i++ {
		upload := uint64(1024 * (10 + i*5)) // Start at 10KB, increase by 5KB each time
		download := uint64(1024 * (5 + i*3)) // Start at 5KB, increase by 3KB each time
		
		c.AddDataPoint(upload, download)
		
		fmt.Printf("\nStep %d: Added upload=%dKB, download=%dKB\n", i+1, upload/1024, download/1024)
		debugWindows(c, i+1)
		
		output := c.Render()
		fmt.Printf("Chart: %s\n", compactChart(output))
	}
}

func debugWindows(c *chart.BrailleChart, dataLen int) {
	windowSize := 3 // For TimeScale3Min
	chartWidth := 10
	
	// Calculate windows using the same logic as renderWithTimeWindows
	totalCompleteWindows := dataLen / windowSize
	hasPartialWindow := (dataLen % windowSize) != 0
	totalWindows := totalCompleteWindows
	if hasPartialWindow {
		totalWindows++
	}
	
	firstVisibleWindow := 0
	if totalWindows > chartWidth {
		firstVisibleWindow = totalWindows - chartWidth
	}
	
	fmt.Printf("  DataLen=%d, CompleteWindows=%d, HasPartial=%v, TotalWindows=%d\n", 
		dataLen, totalCompleteWindows, hasPartialWindow, totalWindows)
	fmt.Printf("  FirstVisible=%d, VisibleWindows=%d\n", 
		firstVisibleWindow, min(totalWindows-firstVisibleWindow, chartWidth))
	
	// Show window mappings
	for x := 0; x < chartWidth && x < totalWindows-firstVisibleWindow; x++ {
		windowIndex := firstVisibleWindow + x
		windowStartIndex := windowIndex * windowSize
		windowEndIndex := windowStartIndex + windowSize
		if windowEndIndex > dataLen {
			windowEndIndex = dataLen
		}
		
		fmt.Printf("  Column %d -> Window %d [%d,%d)\n", 
			x, windowIndex, windowStartIndex, windowEndIndex)
	}
}

func compactChart(chart string) string {
	// Return just the first line to see the pattern
	lines := fmt.Sprintf("%q", chart)
	if len(lines) > 80 {
		return lines[:80] + "..."
	}
	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}