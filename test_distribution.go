package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Testing Time Scale Width Distribution")
	fmt.Println("====================================")

	// Create chart similar to what would be running in the app
	bc := chart.NewBrailleChart(500) // Large buffer
	bc.SetHeight(10)

	// Add substantial data (simulating a real session)
	fmt.Println("Adding 200 data points to simulate real usage...")
	for i := 0; i < 200; i++ {
		// Create some variation in data to make it interesting
		base := uint64(10000)
		variation := uint64(i * 100)
		spike := uint64(0)
		if i%20 == 0 { // Add spikes every 20 points
			spike = uint64(50000)
		}
		
		upload := base + variation + spike
		download := base*8/10 + variation*8/10 + spike*8/10
		bc.AddDataPoint(upload, download)
	}

	// Test the time scales that were showing squashed data
	testScales := []chart.TimeScale{
		chart.TimeScale1Min,
		chart.TimeScale5Min,
		chart.TimeScale15Min,
		chart.TimeScale30Min,
	}

	for _, scale := range testScales {
		bc.SetTimeScale(scale)
		
		fmt.Printf("\n=== %s TIME SCALE ===\n", bc.GetTimeScaleName())
		
		rendered := bc.Render()
		lines := strings.Split(rendered, "\n")
		
		if len(lines) > 0 {
			// Analyze the first line to see data distribution
			firstLine := lines[0]
			
			// Remove ANSI color codes for analysis
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
			
			// Count characters by position segments
			width := len(cleanLine)
			if width > 0 {
				segments := 4 // Divide into 4 segments to check distribution
				segmentSize := width / segments
				
				fmt.Printf("Line width: %d characters\n", width)
				fmt.Printf("Data distribution across %d segments:\n", segments)
				
				for seg := 0; seg < segments; seg++ {
					start := seg * segmentSize
					end := start + segmentSize
					if seg == segments-1 {
						end = width // Last segment takes remainder
					}
					
					nonSpaceCount := 0
					for i := start; i < end && i < len(cleanLine); i++ {
						if cleanLine[i] != ' ' {
							nonSpaceCount++
						}
					}
					
					segment := ""
					if end > len(cleanLine) {
						end = len(cleanLine)
					}
					if start < len(cleanLine) {
						segment = cleanLine[start:end]
					}
					
					fmt.Printf("  Segment %d (chars %d-%d): %d data chars", seg+1, start, end-1, nonSpaceCount)
					if nonSpaceCount == 0 {
						fmt.Printf(" [EMPTY]")
					}
					fmt.Printf("\n")
					
					// Show a sample of the segment
					displaySample := segment
					if len(displaySample) > 20 {
						displaySample = displaySample[:17] + "..."
					}
					fmt.Printf("    Sample: '%s'\n", displaySample)
				}
				
				// Overall check
				totalNonSpace := 0
				for _, char := range cleanLine {
					if char != ' ' {
						totalNonSpace++
					}
				}
				
				dataPercentage := float64(totalNonSpace) / float64(width) * 100
				fmt.Printf("Total data characters: %d (%.1f%% of width)\n", totalNonSpace, dataPercentage)
				
				// Check if data is well distributed (not squashed to one side)
				leftHalf := 0
				rightHalf := 0
				midpoint := width / 2
				
				for i, char := range cleanLine {
					if char != ' ' {
						if i < midpoint {
							leftHalf++
						} else {
							rightHalf++
						}
					}
				}
				
				if leftHalf > 0 && rightHalf > 0 {
					fmt.Printf("✓ Data distributed across chart (left: %d, right: %d)\n", leftHalf, rightHalf)
				} else if leftHalf > 0 {
					fmt.Printf("⚠ Data squashed to LEFT side only (%d chars)\n", leftHalf)
				} else if rightHalf > 0 {
					fmt.Printf("⚠ Data squashed to RIGHT side only (%d chars)\n", rightHalf)
				} else {
					fmt.Printf("⚠ No data visible\n")
				}
			}
		}
	}
}
