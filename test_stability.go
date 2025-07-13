package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Testing Window Stability")
	fmt.Println("========================")

	bc := chart.NewBrailleChart(200)
	bc.SetHeight(6)

	// Add enough data for meaningful windows
	for i := 0; i < 150; i++ {
		bc.AddDataPoint(uint64(1000+i*100), uint64(800+i*80))
	}

	// Test 5-minute scale stability
	bc.SetTimeScale(chart.TimeScale5Min)
	
	fmt.Printf("Testing %s scale:\n", bc.GetTimeScaleName())
	
	// Capture multiple snapshots to see consistency
	snapshots := make([]string, 5)
	for i := 0; i < 5; i++ {
		snapshots[i] = bc.Render()
		// Add one data point between snapshots
		bc.AddDataPoint(uint64(20000+i*1000), uint64(18000+i*800))
	}
	
	// Analyze how much the chart changes between snapshots
	for i := 1; i < len(snapshots); i++ {
		beforeLines := strings.Split(snapshots[i-1], "\n")
		afterLines := strings.Split(snapshots[i], "\n")
		
		changed := 0
		for j := 0; j < len(beforeLines) && j < len(afterLines); j++ {
			if beforeLines[j] != afterLines[j] {
				changed++
			}
		}
		
		fmt.Printf("Snapshot %d->%d: %d lines changed\n", i-1, i, changed)
		
		if changed <= 2 {
			fmt.Printf("  ✓ Stable change\n")
		} else if changed <= 5 {
			fmt.Printf("  ~ Moderate change\n")
		} else {
			fmt.Printf("  ⚠ High instability (%d lines)\n", changed)
		}
	}
	
	// Test that rightmost data is most recent
	lastSnapshot := snapshots[len(snapshots)-1]
	lines := strings.Split(lastSnapshot, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		// Remove ANSI codes
		cleanLine := ""
		inAnsi := false
		for _, char := range firstLine {
			if char == '\x1b' {
				inAnsi = true
				continue
			}
			if inAnsi {
				if char == 'm' {
					inAnsi = false
				}
				continue
			}
			cleanLine += string(char)
		}
		
		// Find rightmost data
		rightmostData := -1
		for i := len(cleanLine) - 1; i >= 0; i-- {
			if cleanLine[i] != ' ' {
				rightmostData = i
				break
			}
		}
		
		if rightmostData >= len(cleanLine)-10 {
			fmt.Printf("✓ Data correctly positioned on right edge\n")
		} else {
			fmt.Printf("⚠ Data not on right edge (position %d of %d)\n", rightmostData, len(cleanLine))
		}
	}
}
