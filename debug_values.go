package main

import (
	"fmt"
	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Value Debugging")
	fmt.Println("===============")
	
	c := chart.NewBrailleChart(1000)
	c.SetWidth(40) 
	c.SetHeight(6)
	c.SetTimeScale(chart.TimeScale3Min)
	
	// Add data points and debug the aggregated values
	dataPoints := []struct{upload, download uint64}{
		{10*1024, 5*1024},   // Step 1
		{15*1024, 8*1024},   // Step 2
		{20*1024, 11*1024},  // Step 3
		{25*1024, 14*1024},  // Step 4
		{30*1024, 17*1024},  // Step 5
	}
	
	for i, data := range dataPoints {
		c.AddDataPoint(data.upload, data.download)
		
		fmt.Printf("\nStep %d: Added upload=%dKB, download=%dKB\n", 
			i+1, data.upload/1024, data.download/1024)
		
		// Calculate what the aggregated values should be manually
		debugWindowAggregation(i+1, dataPoints[:i+1])
		
		// Get the current max value for scaling
		fmt.Printf("Chart maxValue: %d\n", c.GetMaxValue())
	}
}

func debugWindowAggregation(step int, data []struct{upload, download uint64}) {
	windowSize := 3
	dataLen := len(data)
	
	// Calculate windows using the same logic as the chart
	totalCompleteWindows := dataLen / windowSize
	hasPartialWindow := (dataLen % windowSize) != 0
	totalWindows := totalCompleteWindows
	if hasPartialWindow {
		totalWindows++
	}
	
	fmt.Printf("  Data length: %d, Total windows: %d\n", dataLen, totalWindows)
	
	// Show aggregated values for each window
	for windowIndex := 0; windowIndex < totalWindows; windowIndex++ {
		windowStartIndex := windowIndex * windowSize
		windowEndIndex := windowStartIndex + windowSize
		if windowEndIndex > dataLen {
			windowEndIndex = dataLen
		}
		
		var maxUpload, maxDownload uint64
		
		for i := windowStartIndex; i < windowEndIndex; i++ {
			if data[i].upload > maxUpload {
				maxUpload = data[i].upload
			}
			if data[i].download > maxDownload {
				maxDownload = data[i].download
			}
		}
		
		fmt.Printf("  Window %d [%d,%d): upload=%dKB, download=%dKB\n",
			windowIndex, windowStartIndex, windowEndIndex, 
			maxUpload/1024, maxDownload/1024)
	}
}
