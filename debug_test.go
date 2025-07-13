package main

import (
	"fmt"
	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	// Create a chart and add some test data
	c := chart.NewBrailleChart(1000)
	c.SetWidth(10) // Small width for easy debugging
	
	// Add 100 test data points with predictable values
	for i := 0; i < 100; i++ {
		upload := uint64(i * 100)    // 0, 100, 200, 300, etc.
		download := uint64(i * 200)  // 0, 200, 400, 600, etc.
		c.AddDataPoint(upload, download)
	}
	
	fmt.Println("=== 1 MINUTE MODE ===")
	c.SetTimeScale(chart.TimeScale1Min)
	testRendering(c, "1min")
	
	fmt.Println("\n=== 3 MINUTE MODE ===")
	c.SetTimeScale(chart.TimeScale3Min)
	testRendering(c, "3min")
	
	fmt.Println("\n=== 60 MINUTE MODE ===")
	c.SetTimeScale(chart.TimeScale60Min)
	testRendering(c, "60min")
}

func testRendering(c *chart.BrailleChart, mode string) {
	// Simulate what happens in the rendering loop
	dataLen := c.GetDataLength() // We need to add this method
	chartWidth := 10
	
	fmt.Printf("DataLen: %d, ChartWidth: %d\n", dataLen, chartWidth)
	
	// Test the indexing logic for each column
	for x := 0; x < chartWidth; x++ {
		// This is the current logic from braille.go
		stepSize := 1
		if mode != "1min" {
			timeScaleSeconds := c.GetTimeScaleSeconds()
			stepSize = timeScaleSeconds / 60
		}
		
		baseIndex := chartWidth - x
		dataIndex := dataLen - (baseIndex * stepSize)
		
		fmt.Printf("Column %d: stepSize=%d, baseIndex=%d, dataIndex=%d\n", 
			x, stepSize, baseIndex, dataIndex)
	}
}
