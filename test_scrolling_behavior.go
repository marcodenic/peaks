package main

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("REAL-TIME SCROLLING BEHAVIOR TEST")
	fmt.Println("=================================")
	fmt.Println("This test simulates adding data points one by one to see")
	fmt.Println("exactly when and how bar heights change as data scrolls.")
	fmt.Println()
	
	// Test 3-minute time scale (windowSize = 3)
	bc := chart.NewBrailleChart(100)
	bc.SetWidth(8) // Small width to see changes clearly
	bc.SetHeight(6)
	bc.SetTimeScale(chart.TimeScale3Min)
	
	fmt.Printf("Testing %s time scale (windowSize = 3)\n", bc.GetTimeScaleName())
	fmt.Println("Expected behavior: Only rightmost column should change as new data arrives")
	fmt.Println("All other columns should remain EXACTLY the same height")
	fmt.Println()
	
	// Add data points one by one and observe behavior
	testData := []uint64{
		1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000,
		11000, 12000, 13000, 14000, 15000, 16000, 17000, 18000, 19000, 20000,
	}
	
	var previousChart []string
	
	for i, uploadValue := range testData {
		downloadValue := uploadValue / 2 // Some download data
		bc.AddDataPoint(uploadValue, downloadValue)
		
		currentChart := strings.Split(bc.Render(), "\n")
		
		fmt.Printf("--- After adding data point %d (upload: %d) ---\n", i+1, uploadValue)
		
		// Print the chart
		for lineIdx, line := range currentChart {
			if lineIdx < 4 { // Show first 4 lines
				fmt.Printf("  %s\n", line)
			}
		}
		
		// Compare with previous chart to detect changes
		if previousChart != nil && len(previousChart) == len(currentChart) {
			fmt.Print("  Changes from previous: ")
			changesDetected := false
			
			// Check each character position to see what changed
			for lineIdx := 0; lineIdx < 4 && lineIdx < len(currentChart); lineIdx++ {
				currentLine := currentChart[lineIdx]
				previousLine := previousChart[lineIdx]
				
				// Pad lines to same length for comparison
				maxLen := len(currentLine)
				if len(previousLine) > maxLen {
					maxLen = len(previousLine)
				}
				
				for charIdx := 0; charIdx < maxLen; charIdx++ {
					var currentChar, previousChar rune = ' ', ' '
					
					if charIdx < len(currentLine) {
						currentChar = rune(currentLine[charIdx])
					}
					if charIdx < len(previousLine) {
						previousChar = rune(previousLine[charIdx])
					}
					
					if currentChar != previousChar {
						columnNumber := charIdx
						fmt.Printf("[Line%d,Col%d: '%c'→'%c'] ", lineIdx, columnNumber, previousChar, currentChar)
						changesDetected = true
					}
				}
			}
			
			if !changesDetected {
				fmt.Print("NO CHANGES (Good!)")
			}
			fmt.Println()
		}
		
		previousChart = make([]string, len(currentChart))
		copy(previousChart, currentChart)
		
		fmt.Println()
		time.Sleep(100 * time.Millisecond) // Small delay for readability
	}
	
	fmt.Println("=== TEST ANALYSIS ===")
	fmt.Println("✓ GOOD: Only rightmost columns change as new data arrives")
	fmt.Println("✗ BAD: Existing columns change height (bars jump)")
	fmt.Println()
	fmt.Println("If you see existing columns changing, the algorithm needs fixing!")
}
