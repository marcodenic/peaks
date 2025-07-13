package main

import (
	"fmt"
	"regexp"
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
	fmt.Println("Improved Stability Analysis")
	fmt.Println("===========================")
	fmt.Println()

	testTimeScale(chart.TimeScale3Min)
}

func testTimeScale(timeScale chart.TimeScale) {
	width := 40
	height := 6

	c := chart.NewBrailleChart(1000)
	c.SetWidth(width)
	c.SetHeight(height)
	c.SetTimeScale(timeScale)

	// Add data points and capture chart outputs
	var outputs [][]string // Each output is an array of clean lines
	var dataPoints []uint64

	timeScaleMinutes := c.GetTimeScaleSeconds() / 60
	numPoints := timeScaleMinutes + 10

	fmt.Printf("Testing %v time scale (%d minutes window)\n", timeScale, timeScaleMinutes)
	fmt.Printf("Adding %d data points (1 per minute)...\n", numPoints)
	fmt.Println()

	for i := 0; i < numPoints; i++ {
		upload := uint64(1024 * (10 + i*5))
		download := uint64(1024 * (5 + i*3))
		dataPoints = append(dataPoints, upload)

		c.AddDataPoint(upload, download)

		output := c.Render()
		cleanLines := extractCleanLines(output)
		outputs = append(outputs, cleanLines)
	}

	// Now analyze the outputs to detect changes in existing columns
	fmt.Printf("%s=== ANALYSIS ===%s\n", Green, Reset)
	analyzeColumnStability(outputs, dataPoints, timeScale)
}

func extractCleanLines(chart string) []string {
	// Strip ANSI codes
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	clean := ansiRegex.ReplaceAllString(chart, "")
	
	// Split into lines and remove empty ones
	lines := strings.Split(clean, "\n")
	var cleanLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return cleanLines
}

func analyzeColumnStability(outputs [][]string, dataPoints []uint64, timeScale chart.TimeScale) {
	if len(outputs) < 2 {
		fmt.Println("Not enough outputs to analyze")
		return
	}

	changesDetected := 0
	stableCount := 0
	
	for i := 1; i < len(outputs); i++ {
		prev := outputs[i-1]
		curr := outputs[i]
		
		if len(prev) == 0 || len(curr) == 0 {
			continue
		}
		
		// Compare only the overlapping columns
		// The number of columns is determined by the minimum number of non-space characters
		prevCols := countVisibleColumns(prev)
		currCols := countVisibleColumns(curr)
		
		// Only compare the columns that existed in the previous state
		columnsToCompare := prevCols
		if columnsToCompare == 0 {
			continue
		}
		
		existingColumnsChanged := false
		
		// Compare each existing column
		for col := 0; col < columnsToCompare; col++ {
			prevColChars := extractColumnChars(prev, col)
			currColChars := extractColumnChars(curr, col)
			
			// Compare the column characters
			if !compareSlices(prevColChars, currColChars) {
				existingColumnsChanged = true
				break
			}
		}
		
		if existingColumnsChanged {
			changesDetected++
			if changesDetected <= 3 {
				fmt.Printf("%sStep %d -> %d: EXISTING COLUMNS CHANGED!%s\n", Red, i, i+1, Reset)
				fmt.Printf("  Data: %dKB -> %dKB\n", dataPoints[i-1]/1024, dataPoints[i]/1024)
				fmt.Printf("  Previous columns: %d, Current columns: %d\n", prevCols, currCols)
				fmt.Printf("  Previous: %v\n", prev)
				fmt.Printf("  Current:  %v\n", curr)
				fmt.Println()
			}
		} else {
			stableCount++
		}
	}

	// Summary
	totalComparisons := len(outputs) - 1
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
		
		if changesDetected > 3 {
			fmt.Printf("  (Showing only first 3 changes, but %d total changes detected)\n", changesDetected)
		}
	}
}

func countVisibleColumns(lines []string) int {
	if len(lines) == 0 {
		return 0
	}
	
	// Count non-space characters in the first line as an approximation
	count := 0
	for _, char := range lines[0] {
		if char != ' ' {
			count++
		}
	}
	return count
}

func extractColumnChars(lines []string, columnIndex int) []rune {
	var colChars []rune
	
	for _, line := range lines {
		runes := []rune(line)
		colPos := 0
		for i, r := range runes {
			if r != ' ' {
				if colPos == columnIndex {
					colChars = append(colChars, r)
					break
				}
				colPos++
			}
			if i == len(runes)-1 && colPos <= columnIndex {
				// Column doesn't exist in this line
				colChars = append(colChars, ' ')
			}
		}
	}
	
	return colChars
}

func compareSlices(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
