package main

import (
	"fmt"
	"strings"
	"regexp"

	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Cache Verification")
	fmt.Println("==================")
	
	c := chart.NewBrailleChart(1000)
	c.SetWidth(10) // Small width for analysis
	c.SetHeight(6)
	c.SetTimeScale(chart.TimeScale3Min)
	
	var prevOutput string
	
	for i := 0; i < 8; i++ {
		upload := uint64(1024 * (10 + i*5))
		download := uint64(1024 * (5 + i*3))
		
		c.AddDataPoint(upload, download)
		output := c.Render()
		cleanOutput := stripAnsi(output)
		
		fmt.Printf("\nStep %d: Added upload=%dKB, download=%dKB\n", i+1, upload/1024, download/1024)
		
		// Analyze windows
		windowSize := 3
		dataLen := i + 1
		totalCompleteWindows := dataLen / windowSize
		hasPartialWindow := (dataLen % windowSize) != 0
		
		fmt.Printf("  DataLen=%d, CompleteWindows=%d, HasPartial=%v\n", 
			dataLen, totalCompleteWindows, hasPartialWindow)
		
		// Show chart
		lines := strings.Split(cleanOutput, "\n")
		var nonEmptyLines []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		
		fmt.Printf("  Chart: %v\n", nonEmptyLines)
		
		// Compare with previous for changes
		if i > 0 {
			prevLines := strings.Split(stripAnsi(prevOutput), "\n")
			var prevNonEmpty []string
			for _, line := range prevLines {
				if strings.TrimSpace(line) != "" {
					prevNonEmpty = append(prevNonEmpty, line)
				}
			}
			
			if len(prevNonEmpty) > 0 && len(nonEmptyLines) > 0 {
				firstLineChanged := prevNonEmpty[0] != nonEmptyLines[0]
				fmt.Printf("  First line changed: %v\n", firstLineChanged)
				if firstLineChanged {
					fmt.Printf("    Previous: '%s'\n", prevNonEmpty[0])
					fmt.Printf("    Current:  '%s'\n", nonEmptyLines[0])
				}
			}
		}
		
		prevOutput = output
	}
}

func stripAnsi(input string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(input, "")
}
