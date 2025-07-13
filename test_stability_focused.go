package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/peaks/internal/chart"
)

// ANSI color codes for better visualization
const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Reset  = "\033[0m"
)

func main() {
	fmt.Println("Time Scale Stability Analysis")
	fmt.Println("=============================")
	fmt.Println()

	// Test just one problematic time scale
	testTimeScale(chart.TimeScale3Min)
}

func testTimeScale(timeScale chart.TimeScale) {
	width := 40 // Smaller width for easier comparison
	height := 10

	c := chart.NewBrailleChart(1000) // Large capacity for all time scales
	c.SetWidth(width)
	c.SetHeight(height)
	c.SetTimeScale(timeScale)

	// Add data points and capture chart outputs
	var outputs []string
	var dataPoints []uint64

	// For time scale testing, we need to consider how many points fit in the time window
	timeScaleMinutes := c.GetTimeScaleSeconds() / 60
	numPoints := timeScaleMinutes + 10 // Add some extra to ensure scrolling

	fmt.Printf("Testing %v time scale (%d minutes window)\n", timeScale, timeScaleMinutes)
	fmt.Printf("Adding %d data points (1 per minute)...\n", numPoints)
	fmt.Println()

	for i := 0; i < numPoints; i++ {
		// Simulate data point every minute with values that will show on chart
		upload := uint64(1024 * (10 + i*5)) // Start at 10KB, increase by 5KB each time
		download := uint64(1024 * (5 + i*3)) // Start at 5KB, increase by 3KB each time
		dataPoints = append(dataPoints, upload)

		// Add data point
		c.AddDataPoint(upload, download)

		// Render chart
		output := c.Render()
		outputs = append(outputs, output)
	}

	// Now analyze the outputs to detect bar height changes
	fmt.Printf("%s=== ANALYSIS ===%s\n", Green, Reset)
	analyzeChartChanges(outputs, dataPoints, timeScale)
}

func analyzeChartChanges(outputs []string, dataPoints []uint64, timeScale chart.TimeScale) {
	if len(outputs) < 2 {
		fmt.Println("Not enough outputs to analyze")
		return
	}

	// Extract chart data (skip headers/footers)
	chartLines := make([][]string, len(outputs))
	for i, output := range outputs {
		lines := strings.Split(output, "\n")
		// Find the actual chart part (skip headers, take middle section)
		chartStart := -1
		chartEnd := -1
		for j, line := range lines {
			if strings.Contains(line, "┌") || strings.Contains(line, "╭") {
				chartStart = j + 1
			}
			if strings.Contains(line, "└") || strings.Contains(line, "╰") {
				chartEnd = j
				break
			}
		}
		if chartStart >= 0 && chartEnd > chartStart {
			chartLines[i] = lines[chartStart:chartEnd]
		} else if len(lines) > 2 {
			// Fallback: take middle portion
			start := len(lines) / 4
			end := len(lines) - len(lines)/4
			if end > start {
				chartLines[i] = lines[start:end]
			}
		}
	}

	// Compare consecutive charts to detect changes in existing columns
	changesDetected := 0
	stableCount := 0

	for i := 1; i < len(chartLines); i++ {
		if len(chartLines[i-1]) == 0 || len(chartLines[i]) == 0 {
			continue
		}

		// Compare the charts column by column (excluding the rightmost column)
		prev := chartLines[i-1]
		curr := chartLines[i]

		// Ensure both have same height
		minHeight := len(prev)
		if len(curr) < minHeight {
			minHeight = len(curr)
		}

		if minHeight == 0 {
			continue
		}

		// Compare each row, excluding the rightmost few characters
		// (the rightmost column can change as it's still aggregating)
		columnChanges := false
		
		for row := 0; row < minHeight; row++ {
			prevRow := prev[row]
			currRow := curr[row]
			
			// Compare all but the last few characters (rightmost column)
			compareLength := len(prevRow)
			if len(currRow) < compareLength {
				compareLength = len(currRow)
			}
			
			// Exclude rightmost 5 characters to allow for rightmost column changes
			if compareLength > 5 {
				compareLength -= 5
			}

			if compareLength > 0 {
				prevPart := prevRow[:compareLength]
				currPart := currRow[:compareLength]
				
				if prevPart != currPart {
					columnChanges = true
					break
				}
			}
		}

		if columnChanges {
			changesDetected++
			fmt.Printf("%sStep %d -> %d: EXISTING COLUMNS CHANGED!%s\n", Red, i, i+1, Reset)
			fmt.Printf("  Data: %dKB -> %dKB\n", dataPoints[i-1]/1024, dataPoints[i]/1024)
			
			// Show the difference
			fmt.Println("  Previous chart:")
			for _, line := range prev {
				fmt.Printf("    '%s'\n", line)
			}
			fmt.Println("  Current chart:")
			for _, line := range curr {
				fmt.Printf("    '%s'\n", line)
			}
			fmt.Println()
		} else {
			stableCount++
		}
	}

	// Summary
	totalComparisons := len(chartLines) - 1
	if totalComparisons > 0 {
		fmt.Printf("Summary for %v time scale:\n", timeScale)
		fmt.Printf("  Total step comparisons: %d\n", totalComparisons)
		fmt.Printf("  %sChanges detected: %d%s\n", 
			func() string { if changesDetected > 0 { return Red } else { return Green } }(),
			changesDetected, Reset)
		fmt.Printf("  %sStable transitions: %d%s\n", Green, stableCount, Reset)
		
		if changesDetected == 0 {
			fmt.Printf("  %s✓ PERFECT: No existing columns changed during scrolling%s\n", Green, Reset)
		} else {
			fmt.Printf("  %s✗ PROBLEM: Existing columns changed %d times during scrolling%s\n", Red, changesDetected, Reset)
			stabilityPercent := float64(stableCount) / float64(totalComparisons) * 100
			fmt.Printf("  Stability: %.1f%% (%d/%d stable transitions)\n", stabilityPercent, stableCount, totalComparisons)
		}
	}
}
