package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Simple Chart Test")
	fmt.Println("=================")
	
	// Create chart
	c := chart.NewBrailleChart(100)
	c.SetWidth(40)
	c.SetHeight(10)
	
	// Test with larger values to ensure they show up
	values := []uint64{
		1024 * 10, // 10KB
		1024 * 20, // 20KB
		1024 * 30, // 30KB
		1024 * 40, // 40KB
		1024 * 50, // 50KB
	}
	
	for i, upload := range values {
		download := upload / 2 // Half the upload speed
		fmt.Printf("Adding data point %d: upload=%d, download=%d\n", i+1, upload, download)
		
		c.AddDataPoint(upload, download)
		output := c.Render()
		
		fmt.Println("Chart output:")
		lines := strings.Split(output, "\n")
		for j, line := range lines {
			if j < len(lines)-1 { // Skip empty last line
				fmt.Printf("  %s\n", line)
			}
		}
		fmt.Println()
	}
}
