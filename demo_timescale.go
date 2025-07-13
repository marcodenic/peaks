package main

import (
	"fmt"
	"time"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Interactive Time Scale Demo")
	fmt.Println("==========================")
	
	// Create chart and add some data
	bc := chart.NewBrailleChart(60)
	bc.SetHeight(12)
	
	// Add data with a clear pattern
	fmt.Println("Adding 200 data points with increasing pattern...")
	for i := 0; i < 200; i++ {
		upload := uint64(1000 + i*50)
		download := uint64(800 + i*40)
		bc.AddDataPoint(upload, download)
	}
	
	// Test all time scales
	timeScales := []chart.TimeScale{
		chart.TimeScale1Min,
		chart.TimeScale3Min,
		chart.TimeScale5Min,
		chart.TimeScale10Min,
		chart.TimeScale15Min,
		chart.TimeScale30Min,
		chart.TimeScale60Min,
	}
	
	for _, scale := range timeScales {
		bc.SetTimeScale(scale)
		fmt.Printf("\n=== %s ===\n", bc.GetTimeScaleName())
		fmt.Printf("Window size: %d data points per column\n", bc.GetTimeScaleSeconds()/60)
		
		// Render chart
		rendered := bc.Render()
		lines := splitLines(rendered)
		
		// Show only first few lines and last few lines for compact display
		fmt.Println("Chart preview (first 3 and last 3 lines):")
		for i := 0; i < len(lines) && i < 3; i++ {
			fmt.Printf("  %s\n", lines[i])
		}
		if len(lines) > 6 {
			fmt.Println("  ...")
		}
		for i := len(lines) - 3; i < len(lines) && i >= 3; i++ {
			fmt.Printf("  %s\n", lines[i])
		}
		
		// Add one more data point and show stability
		fmt.Println("\nAdding one more data point...")
		oldRender := rendered
		bc.AddDataPoint(uint64(20000), uint64(18000))
		newRender := bc.Render()
		
		changedLines := countChangedLines(oldRender, newRender)
		fmt.Printf("Lines changed: %d (lower is better for stability)\n", changedLines)
		
		if scale != chart.TimeScale1Min && changedLines > 2 {
			fmt.Printf("⚠ WARNING: Too many lines changed for %s scale!\n", bc.GetTimeScaleName())
		} else if scale != chart.TimeScale1Min && changedLines <= 2 {
			fmt.Printf("✓ GOOD: Chart is stable for %s scale\n", bc.GetTimeScaleName())
		}
		
		time.Sleep(time.Millisecond * 500) // Brief pause for readability
	}
	
	fmt.Println("\n=== Summary ===")
	fmt.Println("The time scale feature allows you to view different time windows:")
	fmt.Println("- 1 minute: Shows recent data with 1:1 scrolling")
	fmt.Println("- 3+ minutes: Shows aggregated data with stable windows")
	fmt.Println("- Press 't' key in the app to cycle through time scales")
	fmt.Println("- Each column shows the maximum bandwidth in its time window")
}

func splitLines(text string) []string {
	lines := []string{}
	current := ""
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func countChangedLines(before, after string) int {
	beforeLines := splitLines(before)
	afterLines := splitLines(after)
	
	maxLines := len(beforeLines)
	if len(afterLines) > maxLines {
		maxLines = len(afterLines)
	}
	
	changed := 0
	for i := 0; i < maxLines; i++ {
		var beforeLine, afterLine string
		if i < len(beforeLines) {
			beforeLine = beforeLines[i]
		}
		if i < len(afterLines) {
			afterLine = afterLines[i]
		}
		
		if beforeLine != afterLine {
			changed++
		}
	}
	
	return changed
}
