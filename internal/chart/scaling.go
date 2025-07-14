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
	if bc.scalingMode != mode {
		bc.scalingMode = mode
		// Invalidate column cache since scaling affects rendering
		bc.invalidateColumnCache()
	}
}

// GetScalingMode returns the current scaling mode
func (bc *BrailleChart) GetScalingMode() ScalingMode {
	return bc.scalingMode
}

// CycleScalingMode cycles through available scaling modes
func (bc *BrailleChart) CycleScalingMode() ScalingMode {
	oldMode := bc.scalingMode
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
	
	// Invalidate column cache if mode changed
	if oldMode != bc.scalingMode {
		bc.invalidateColumnCache()
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

// GetTimeScale returns the current time scale
func (bc *BrailleChart) GetTimeScale() TimeScale {
	return bc.timeScale
}

// CycleTimeScale cycles through available time scales
func (bc *BrailleChart) CycleTimeScale() TimeScale {
	oldScale := bc.timeScale
	switch bc.timeScale {
	case TimeScale1Min:
		bc.timeScale = TimeScale3Min
	case TimeScale3Min:
		bc.timeScale = TimeScale5Min
	case TimeScale5Min:
		bc.timeScale = TimeScale10Min
	case TimeScale10Min:
		bc.timeScale = TimeScale15Min
	case TimeScale15Min:
		bc.timeScale = TimeScale30Min
	case TimeScale30Min:
		bc.timeScale = TimeScale60Min
	case TimeScale60Min:
		bc.timeScale = TimeScale1Min
	default:
		bc.timeScale = TimeScale1Min
	}
	
	// Invalidate column cache if time scale changed (different aggregation)
	if oldScale != bc.timeScale {
		bc.invalidateColumnCache()
	}
	
	return bc.timeScale
}

// GetTimeScaleName returns a human-readable name for the current time scale
func (bc *BrailleChart) GetTimeScaleName() string {
	switch bc.timeScale {
	case TimeScale1Min:
		return "1m"
	case TimeScale3Min:
		return "3m"
	case TimeScale5Min:
		return "5m"
	case TimeScale10Min:
		return "10m"
	case TimeScale15Min:
		return "15m"
	case TimeScale30Min:
		return "30m"
	case TimeScale60Min:
		return "60m"
	default:
		return "1m"
	}
}

// GetTimeScaleSeconds returns the number of seconds for the current time scale
func (bc *BrailleChart) GetTimeScaleSeconds() int {
	switch bc.timeScale {
	case TimeScale1Min:
		return 60
	case TimeScale3Min:
		return 180
	case TimeScale5Min:
		return 300
	case TimeScale10Min:
		return 600
	case TimeScale15Min:
		return 900
	case TimeScale30Min:
		return 1800
	case TimeScale60Min:
		return 3600
	default:
		return 60
	}
}

// GetTimeScaleMaxPoints calculates the maximum data points needed for the current time scale
// Given that data is collected every 500ms (0.5 seconds), we need 2 points per second
func (bc *BrailleChart) GetTimeScaleMaxPoints() int {
	return bc.GetTimeScaleSeconds() * 2 // 2 data points per second (500ms intervals)
}
