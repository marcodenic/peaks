package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Testing Chart Width Fill")
	fmt.Println("========================")

	// Create chart with known width
	bc := chart.NewBrailleChart(60)
	bc.SetHeight(8)

	// Add some data
	for i := 0; i < 50; i++ {
		upload := uint64(1000 + i*100)
		download := uint64(800 + i*80)
		bc.AddDataPoint(upload, download)
	}

	// Test different time scales
	timeScales := []chart.TimeScale{
		chart.TimeScale1Min,
		chart.TimeScale5Min,
		chart.TimeScale15Min,
		chart.TimeScale60Min,
	}

	for _, scale := range timeScales {
		bc.SetTimeScale(scale)
		rendered := bc.Render()
		lines := strings.Split(rendered, "\n")
		
		fmt.Printf("\n=== %s ===\n", bc.GetTimeScaleName())
		fmt.Printf("Chart width should be: %d\n", bc.GetWidth())
		
		// Check each line length
		for i, line := range lines {
			// Remove ANSI color codes for accurate length measurement
			cleanLine := removeAnsiCodes(line)
			fmt.Printf("Line %d length: %d chars", i, len(cleanLine))
			if len(cleanLine) < bc.GetWidth() {
				fmt.Printf(" ⚠ TOO SHORT! (expected %d)\n", bc.GetWidth())
			} else if len(cleanLine) > bc.GetWidth() {
				fmt.Printf(" ⚠ TOO LONG! (expected %d)\n", bc.GetWidth())
			} else {
				fmt.Printf(" ✓ GOOD\n")
			}
		}
		
		// Show first line to see distribution
		if len(lines) > 0 {
			cleanFirstLine := removeAnsiCodes(lines[0])
			fmt.Printf("First line content: '%s'\n", cleanFirstLine)
			
			// Count non-space characters to see data distribution
			nonSpaceCount := 0
			for _, char := range cleanFirstLine {
				if char != ' ' {
					nonSpaceCount++
				}
			}
			fmt.Printf("Non-space characters: %d (data distribution)\n", nonSpaceCount)
		}
	}
}

// removeAnsiCodes removes ANSI color codes from a string for accurate length measurement
func removeAnsiCodes(s string) string {
	result := ""
	inEscape := false
	
	for _, char := range s {
		if char == '\x1b' { // ESC character
			inEscape = true
			continue
		}
		if inEscape {
			if char == 'm' { // End of ANSI sequence
				inEscape = false
			}
			continue
		}
		result += string(char)
	}
	
	return result
}
