package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	// Create a chart with a small width for easier testing
	bc := chart.NewBrailleChart(20) // 20 columns wide
	bc.SetHeight(6) // 6 rows tall for easier visual comparison

	fmt.Println("Testing Time Scale Stability - Visual Test")
	fmt.Println("==========================================")

	// Add some data with a clear pattern
	for i := 0; i < 100; i++ {
		upload := uint64(1000 + i*100)   // Increasing pattern
		download := uint64(800 + i*80)   // Different increasing pattern
		bc.AddDataPoint(upload, download)
	}

	// Test 1-minute vs 5-minute time scales
	fmt.Println("\n=== 1-MINUTE TIME SCALE TEST ===")
	bc.SetTimeScale(chart.TimeScale1Min)
	
	fmt.Printf("Chart width: %d\n", bc.GetWidth())
	fmt.Printf("Data points: %d\n", bc.GetDataLength())
	
	// Render initial state
	beforeRender := bc.Render()
	fmt.Println("BEFORE adding new data:")
	fmt.Println(beforeRender)
	
	// Add one data point
	bc.AddDataPoint(uint64(20000), uint64(18000))
	
	// Render after state
	afterRender := bc.Render()
	fmt.Println("\nAFTER adding new data:")
	fmt.Println(afterRender)
	
	// Compare line by line
	fmt.Println("\nLINE-BY-LINE COMPARISON (1-minute scale):")
	compareRenderings(beforeRender, afterRender)

	fmt.Println("\n=== 5-MINUTE TIME SCALE TEST ===")
	bc.SetTimeScale(chart.TimeScale5Min)
	
	// Render initial state
	beforeRender = bc.Render()
	fmt.Println("BEFORE adding new data:")
	fmt.Println(beforeRender)
	
	// Add one data point
	bc.AddDataPoint(uint64(21000), uint64(19000))
	
	// Render after state
	afterRender = bc.Render()
	fmt.Println("\nAFTER adding new data:")
	fmt.Println(afterRender)
	
	// Compare line by line
	fmt.Println("\nLINE-BY-LINE COMPARISON (5-minute scale):")
	compareRenderings(beforeRender, afterRender)
	
	fmt.Println("\n=== WINDOW SIZE CALCULATION TEST ===")
	for _, scale := range []chart.TimeScale{
		chart.TimeScale1Min,
		chart.TimeScale3Min,
		chart.TimeScale5Min,
		chart.TimeScale10Min,
		chart.TimeScale15Min,
		chart.TimeScale30Min,
		chart.TimeScale60Min,
	} {
		bc.SetTimeScale(scale)
		timeScaleSeconds := bc.GetTimeScaleSeconds()
		windowSize := timeScaleSeconds / 60
		if windowSize < 1 {
			windowSize = 1
		}
		fmt.Printf("%s: %d seconds, window size: %d data points per column\n", 
			bc.GetTimeScaleName(), timeScaleSeconds, windowSize)
	}
}

func compareRenderings(before, after string) {
	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")
	
	maxLines := len(beforeLines)
	if len(afterLines) > maxLines {
		maxLines = len(afterLines)
	}
	
	changedLines := 0
	
	for i := 0; i < maxLines; i++ {
		var beforeLine, afterLine string
		if i < len(beforeLines) {
			beforeLine = beforeLines[i]
		}
		if i < len(afterLines) {
			afterLine = afterLines[i]
		}
		
		if beforeLine != afterLine {
			changedLines++
			fmt.Printf("  Line %d CHANGED:\n", i)
			fmt.Printf("    Before: '%s'\n", beforeLine)
			fmt.Printf("    After:  '%s'\n", afterLine)
			
			// Show character-by-character differences
			minLen := len(beforeLine)
			if len(afterLine) < minLen {
				minLen = len(afterLine)
			}
			
			changes := 0
			for j := 0; j < minLen; j++ {
				if beforeLine[j] != afterLine[j] {
					changes++
				}
			}
			
			if len(beforeLine) != len(afterLine) {
				changes += abs(len(beforeLine) - len(afterLine))
			}
			
			fmt.Printf("    Character changes: %d\n", changes)
		}
	}
	
	if changedLines == 0 {
		fmt.Println("  ✓ No lines changed - chart is stable!")
	} else {
		fmt.Printf("  ⚠ %d lines changed\n", changedLines)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
