package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Character-by-Character Analysis")
	fmt.Println("===============================")
	
	c := chart.NewBrailleChart(1000)
	c.SetWidth(40) 
	c.SetHeight(6)
	c.SetTimeScale(chart.TimeScale3Min)
	
	var prevChart string
	
	// Focus on the critical transition (step 3 -> 4)
	for i := 0; i < 5; i++ {
		upload := uint64(1024 * (10 + i*5))
		download := uint64(1024 * (5 + i*3))
		
		c.AddDataPoint(upload, download)
		currChart := c.Render()
		
		fmt.Printf("\nStep %d: Added upload=%dKB, download=%dKB\n", i+1, upload/1024, download/1024)
		
		// Extract just the chart part
		currLines := extractChartLines(currChart)
		fmt.Printf("Chart lines (%d):\n", len(currLines))
		for j, line := range currLines {
			fmt.Printf("  [%d]: '%s'\n", j, line)
		}
		
		if i > 0 && i == 3 { // Step 3->4 transition
			fmt.Println("\n=== DETAILED COMPARISON (Step 3 vs 4) ===")
			prevLines := extractChartLines(prevChart)
			
			if len(prevLines) > 0 && len(currLines) > 0 {
				fmt.Printf("Previous first line: '%s' (len=%d)\n", prevLines[0], len(prevLines[0]))
				fmt.Printf("Current first line:  '%s' (len=%d)\n", currLines[0], len(currLines[0]))
				
				compareCharByChar(prevLines[0], currLines[0])
			}
		}
		
		prevChart = currChart
	}
}

func extractChartLines(chart string) []string {
	lines := strings.Split(chart, "\n")
	var chartLines []string
	
	// Skip empty lines and find the actual chart content
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			chartLines = append(chartLines, line)
		}
	}
	
	return chartLines
}

func compareCharByChar(prev, curr string) {
	maxLen := len(prev)
	if len(curr) > maxLen {
		maxLen = len(curr)
	}
	
	fmt.Println("Character-by-character comparison:")
	differences := 0
	
	for i := 0; i < maxLen; i++ {
		var prevChar, currChar rune = ' ', ' '
		var prevStr, currStr string = " ", " "
		
		if i < len(prev) {
			prevRunes := []rune(prev)
			if i < len(prevRunes) {
				prevChar = prevRunes[i]
				prevStr = string(prevChar)
			}
		}
		
		if i < len(curr) {
			currRunes := []rune(curr)
			if i < len(currRunes) {
				currChar = currRunes[i]
				currStr = string(currChar)
			}
		}
		
		if prevChar != currChar {
			differences++
			fmt.Printf("  Pos %d: '%s' -> '%s' (DIFF)\n", i, prevStr, currStr)
			if differences > 10 { // Limit output
				fmt.Printf("  ... (%d more differences)\n", maxLen-i-1)
				break
			}
		} else if i < 10 || differences > 0 { // Show first few chars and any around differences
			fmt.Printf("  Pos %d: '%s' -> '%s' (same)\n", i, prevStr, currStr)
		}
	}
	
	fmt.Printf("Total differences: %d\n", differences)
}
