package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Quick Stability Test")
	fmt.Println("==================")

	bc := chart.NewBrailleChart(200)
	bc.SetHeight(6)

	// Add data
	for i := 0; i < 100; i++ {
		bc.AddDataPoint(uint64(1000+i*100), uint64(800+i*80))
	}

	// Test 5-minute scale
	bc.SetTimeScale(chart.TimeScale5Min)
	before := bc.Render()
	
	// Add one more data point
	bc.AddDataPoint(uint64(20000), uint64(18000))
	after := bc.Render()
	
	// Count changed lines
	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")
	
	changed := 0
	for i := 0; i < len(beforeLines) && i < len(afterLines); i++ {
		if beforeLines[i] != afterLines[i] {
			changed++
		}
	}
	
	fmt.Printf("Lines changed: %d\n", changed)
	if changed <= 2 {
		fmt.Println("✓ Chart is stable!")
	} else {
		fmt.Println("⚠ Chart is unstable!")
	}
}
