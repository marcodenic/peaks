// Package chart provides braille chart rendering functionality
package chart

import "github.com/charmbracelet/lipgloss"

const (
	// Chart configuration constants
	MinChartHeight = 8                 // Minimum chart height in rows
	brailleDots    = 4                 // Braille has 4 vertical dots per character
	brailleBase    = 0x2800            // Base braille character code
	maxScaleLimit  = 100 * 1024 * 1024 // 100MB/s maximum scale

	// Optimization: pre-calculated constants
	maxBrailleChars    = 256 // Maximum number of braille characters (0x2800-0x28FF)
	defaultChartWidth  = 80
	defaultChartHeight = 20
	defaultMaxPoints   = 50

	// Scaling constants
	logBase     = 10.0   // Base for logarithmic scaling
	minLogValue = 1024.0 // Minimum value for log scaling (1KB)
)

// ScalingMode defines how the chart scales data
type ScalingMode int

const (
	ScalingLinear ScalingMode = iota
	ScalingLogarithmic
	ScalingSquareRoot
)

// ColorGradient represents a color gradient configuration
type ColorGradient struct {
	Steps []lipgloss.Color
}

// DataPoint represents a single measurement point
type DataPoint struct {
	Upload   uint64
	Download uint64
}

// Optimization: pre-calculated dot patterns as package constants
var dotPatterns = [4]int{
	0x01 | 0x08, // dots 0,3 (row 0)
	0x02 | 0x10, // dots 1,4 (row 1)
	0x04 | 0x20, // dots 2,5 (row 2)
	0x40 | 0x80, // dots 6,7 (row 3)
}
