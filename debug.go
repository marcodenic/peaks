package main

import (
	"fmt"
	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	// Create a chart and add some test data
	c := chart.NewBrailleChart(1000)
	c.SetWidth(10) // Small width for easy debugging
	
	// Add 50 test data points first
	for i := 0; i < 50; i++ {
		upload := uint64(i * 100)    // 0, 100, 200, 300, etc.
		download := uint64(i * 200)  // 0, 200, 400, 600, etc.
		c.AddDataPoint(upload, download)
	}
	
	fmt.Println("=== BEFORE ADDING NEW DATA POINT ===")
	c.SetTimeScale(chart.TimeScale3Min)
	indices1 := getColumnIndices(c, "3min")
	
	// Add one more data point
	c.AddDataPoint(5000, 10000)
	
	fmt.Println("\n=== AFTER ADDING NEW DATA POINT ===")
	indices2 := getColumnIndices(c, "3min")
	
	fmt.Println("\n=== COMPARISON ===")
	for i := 0; i < len(indices1); i++ {
		diff := indices2[i] - indices1[i]
		fmt.Printf("Column %d: %d -> %d (diff: %d)\n", i, indices1[i], indices2[i], diff)
	}
}

func getColumnIndices(c *chart.BrailleChart, mode string) []int {
	dataLen := c.GetDataLength()
	chartWidth := 10
	indices := make([]int, chartWidth)
	
	for x := 0; x < chartWidth; x++ {
		stepSize := 1
		if mode != "1min" {
			timeScaleSeconds := c.GetTimeScaleSeconds()
			stepSize = timeScaleSeconds / 60
		}
		
		baseIndex := chartWidth - x
		dataIndex := dataLen - (baseIndex * stepSize)
		indices[x] = dataIndex
	}
	return indices
}
