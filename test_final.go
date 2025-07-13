package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Testing Scrolling Stability")
	fmt.Println("===========================")

	bc := chart.NewBrailleChart(200)
	bc.SetHeight(6)

	// Add sufficient data
	for i := 0; i < 100; i++ {
		bc.AddDataPoint(uint64(1000+i*100), uint64(800+i*80))
	}

	// Test 5-minute scale stability
	bc.SetTimeScale(chart.TimeScale5Min)
	
	fmt.Printf("Testing %s scale:\n", bc.GetTimeScaleName())
	
	// Capture before state
	before := bc.Render()
	
	// Add one new data point
	bc.AddDataPoint(uint64(20000), uint64(18000))
	
	// Capture after state
	after := bc.Render()
	
	// Count changes
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
		fmt.Printf("✓ Chart is stable! Minimal changes when new data arrives.\n")
	} else {
		fmt.Printf("⚠ Chart is unstable. Too many changes (%d lines).\n", changed)
	}
	
	// Also check that data is on the right edge
	if len(afterLines) > 0 {
		firstLine := afterLines[0]
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
			fmt.Printf("✓ Data appears near right edge (position %d of %d)\n", rightmostData, len(cleanLine))
		} else {
			fmt.Printf("⚠ Data not near right edge (position %d of %d)\n", rightmostData, len(cleanLine))
		}
	}
}
