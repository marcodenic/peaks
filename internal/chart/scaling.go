// Package chart provides scaling functionality for braille charts
package chart

import "math"

// scaleValue applies the current scaling mode to a value
func (bc *BrailleChart) scaleValue(value uint64, maxValue uint64) float64 {
	if value == 0 {
		return 0
	}

	switch bc.scalingMode {
	case ScalingLinear:
		return float64(value) / float64(maxValue)

	case ScalingLogarithmic:
		// Ensure minimum value for log scaling
		val := math.Max(float64(value), minLogValue)
		maxVal := math.Max(float64(maxValue), minLogValue)

		// Apply logarithmic scaling
		logVal := math.Log10(val)
		logMax := math.Log10(maxVal)
		logMin := math.Log10(minLogValue)

		// Normalize to 0-1 range
		return (logVal - logMin) / (logMax - logMin)

	case ScalingSquareRoot:
		return math.Sqrt(float64(value)) / math.Sqrt(float64(maxValue))

	default:
		return float64(value) / float64(maxValue)
	}
}

// SetScalingMode sets the scaling mode for the chart
func (bc *BrailleChart) SetScalingMode(mode ScalingMode) {
	bc.scalingMode = mode
}

// GetScalingMode returns the current scaling mode
func (bc *BrailleChart) GetScalingMode() ScalingMode {
	return bc.scalingMode
}

// CycleScalingMode cycles through available scaling modes
func (bc *BrailleChart) CycleScalingMode() ScalingMode {
	switch bc.scalingMode {
	case ScalingLinear:
		bc.scalingMode = ScalingLogarithmic
	case ScalingLogarithmic:
		bc.scalingMode = ScalingSquareRoot
	case ScalingSquareRoot:
		bc.scalingMode = ScalingLinear
	default:
		bc.scalingMode = ScalingLinear
	}
	return bc.scalingMode
}

// GetScalingModeName returns a human-readable name for the current scaling mode
func (bc *BrailleChart) GetScalingModeName() string {
	switch bc.scalingMode {
	case ScalingLinear:
		return "Linear"
	case ScalingLogarithmic:
		return "Logarithmic"
	case ScalingSquareRoot:
		return "Square Root"
	default:
		return "Unknown"
	}
}
