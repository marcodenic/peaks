package main

import (
	"fmt"
	"strings"
	"github.com/marcodenic/peaks/internal/chart"
)

func main() {
	fmt.Println("Cache Debug")
	fmt.Println("===========")
	
	c := chart.NewBrailleChart(1000)
	c.SetWidth(10) 
	c.SetHeight(6)
	c.SetTimeScale(chart.TimeScale3Min)
	
	// Add data points step by step, focusing on the caching boundary
	for i := 0; i < 8; i++ {
		upload := uint64(1024 * (10 + i*5))
		download := uint64(1024 * (5 + i*3))
		
		c.AddDataPoint(upload, download)
		
		windowSize := 3
		dataLen := i + 1
		totalCompleteWindows := dataLen / windowSize
		
		fmt.Printf("\nStep %d: Added upload=%dKB, download=%dKB\n", i+1, upload/1024, download/1024)
		fmt.Printf("  DataLen=%d, CompleteWindows=%d\n", dataLen, totalCompleteWindows)
		
		// This should trigger the cache update when windows become complete
		output := c.Render()
		fmt.Printf("  Chart first line: '%s'\n", getFirstLine(output))
		
		// Add debug info about caching
		if totalCompleteWindows > 0 {
			fmt.Printf("  -> Window 0 should be cached (complete)\n")
		}
		if dataLen >= 6 && totalCompleteWindows >= 2 {
			fmt.Printf("  -> Window 1 should be cached (complete)\n")
		}
	}
}

func getFirstLine(output string) string {
	lines := []string{}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(strings.ReplaceAll(line, "\x1b", "")) != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}
