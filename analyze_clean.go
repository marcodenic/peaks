package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Clean Character Analysis")
	fmt.Println("========================")
	
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
		
		// Extract and clean chart lines
		currLines := extractAndCleanChartLines(currChart)
		fmt.Printf("Clean chart lines (%d):\n", len(currLines))
		for j, line := range currLines {
			fmt.Printf("  [%d]: '%s'\n", j, line)
		}
		
		if i > 0 && i == 3 { // Step 3->4 transition
			fmt.Println("\n=== DETAILED COMPARISON (Step 3 vs 4) ===")
			prevLines := extractAndCleanChartLines(prevChart)
			
			if len(prevLines) > 0 && len(currLines) > 0 {
				fmt.Printf("Previous first line: '%s' (len=%d)\n", prevLines[0], len(prevLines[0]))
				fmt.Printf("Current first line:  '%s' (len=%d)\n", currLines[0], len(currLines[0]))
				
				compareCleanChars(prevLines[0], currLines[0])
			}
		}
		
		prevChart = currChart
	}
}

func extractAndCleanChartLines(chart string) []string {
	lines := strings.Split(chart, "\n")
	var chartLines []string
	
	// Strip ANSI codes and find non-empty lines
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	
	for _, line := range lines {
		cleaned := ansiRegex.ReplaceAllString(line, "")
		if strings.TrimSpace(cleaned) != "" {
			chartLines = append(chartLines, cleaned)
		}
	}
	
	return chartLines
}

func compareCleanChars(prev, curr string) {
	fmt.Println("Clean character-by-character comparison:")
	
	prevRunes := []rune(strings.TrimSpace(prev))
	currRunes := []rune(strings.TrimSpace(curr))
	
	fmt.Printf("Previous chars: %v\n", prevRunes)
	fmt.Printf("Current chars:  %v\n", currRunes)
	
	maxLen := len(prevRunes)
	if len(currRunes) > maxLen {
		maxLen = len(currRunes)
	}
	
	differences := 0
	for i := 0; i < maxLen; i++ {
		var prevChar, currChar rune = ' ', ' '
		
		if i < len(prevRunes) {
			prevChar = prevRunes[i]
		}
		if i < len(currRunes) {
			currChar = currRunes[i]
		}
		
		if prevChar != currChar {
			differences++
			fmt.Printf("  Pos %d: '%s' (U+%04X) -> '%s' (U+%04X) (DIFF)\n", 
				i, string(prevChar), prevChar, string(currChar), currChar)
		} else if i < 5 { // Show first few
			fmt.Printf("  Pos %d: '%s' (U+%04X) (same)\n", i, string(prevChar), prevChar)
		}
	}
	
	fmt.Printf("Total differences: %d\n", differences)
}
