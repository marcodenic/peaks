package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Testing Right-Edge Scrolling")
	fmt.Println("============================")

	bc := chart.NewBrailleChart(200)
	bc.SetWidth(20) // Small width for easy visualization
	bc.SetHeight(4)

	// Add some data
	for i := 0; i < 30; i++ {
		upload := uint64(1000 + i*100)
		download := uint64(800 + i*80)
		bc.AddDataPoint(upload, download)
	}

	fmt.Printf("Chart width: %d\n", bc.GetWidth())
	fmt.Printf("Data points: %d\n", bc.GetDataLength())

	// Test different time scales
	timeScales := []chart.TimeScale{
		chart.TimeScale1Min,
		chart.TimeScale5Min,
		chart.TimeScale15Min,
	}

	for _, scale := range timeScales {
		bc.SetTimeScale(scale)
		rendered := bc.Render()
		lines := strings.Split(rendered, "\n")
		
		fmt.Printf("\n=== %s ===\n", bc.GetTimeScaleName())
		
		if len(lines) > 0 {
			firstLine := lines[0]
			
			// Remove ANSI codes for analysis
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
			
			fmt.Printf("Raw line: '%s'\n", cleanLine)
			
			// Check where data appears
			firstDataPos := -1
			lastDataPos := -1
			
			for i, char := range cleanLine {
				if char != ' ' {
					if firstDataPos == -1 {
						firstDataPos = i
					}
					lastDataPos = i
				}
			}
			
			width := len(cleanLine)
			if firstDataPos != -1 {
				fmt.Printf("Data positions: %d to %d (out of %d columns)\n", firstDataPos, lastDataPos, width)
				
				if lastDataPos >= width-5 {
					fmt.Printf("✓ Data appears near RIGHT edge (scrolling correctly)\n")
				} else if firstDataPos <= 5 {
					fmt.Printf("⚠ Data appears near LEFT edge (wrong - should scroll from right)\n")
				} else {
					fmt.Printf("? Data appears in MIDDLE\n")
				}
			} else {
				fmt.Printf("No visible data\n")
			}
		}
	}
}
