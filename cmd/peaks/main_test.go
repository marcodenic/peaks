package main

import (
	"testing"
	"time"

	"github.com/marcodenic/peaks/internal/chart"
	"github.com/marcodenic/peaks/internal/monitor"
	"github.com/marcodenic/peaks/internal/ui"
)

func TestNewBandwidthMonitor(t *testing.T) {
	m := monitor.NewBandwidthMonitor()
	if m == nil {
		t.Fatal("NewBandwidthMonitor returned nil")
	}

	// Test getting rates (should not error)
	_, _, err := m.GetCurrentRates()
	if err != nil {
		t.Logf("GetCurrentRates error (expected on some systems): %v", err)
	}
}

func TestNewBrailleChart(t *testing.T) {
	c := chart.NewBrailleChart(100)
	if c == nil {
		t.Fatal("NewBrailleChart returned nil")
	}

	// Test adding data points
	c.AddDataPoint(1024, 2048)
	c.AddDataPoint(2048, 4096)

	// Test rendering
	output := c.Render()
	if output == "" {
		t.Fatal("Chart render returned empty string")
	}

	// Test reset
	c.Reset()
}

func TestUIComponents(t *testing.T) {
	components := ui.NewComponents()
	if components == nil {
		t.Fatal("NewComponents returned nil")
	}

	stats := components.GetStats()
	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	// Test updating stats
	stats.Update(1024, 2048)
	if stats.PeakUpload != 1024 {
		t.Errorf("Expected PeakUpload 1024, got %d", stats.PeakUpload)
	}
	if stats.PeakDownload != 2048 {
		t.Errorf("Expected PeakDownload 2048, got %d", stats.PeakDownload)
	}
}

func TestFormatters(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{1024, "1.00 KB/s"},
		{1048576, "1.00 MB/s"},
		{1073741824, "1.00 GB/s"},
	}

	for _, test := range tests {
		result := ui.FormatBandwidth(test.input)
		if result != test.expected {
			t.Errorf("FormatBandwidth(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}

	// Test duration formatting
	duration := 125 * time.Second
	result := ui.FormatDuration(duration)
	expected := "2m5s"
	if result != expected {
		t.Errorf("FormatDuration(%v) = %s, expected %s", duration, result, expected)
	}
}

func TestKeyMap(t *testing.T) {
	keys := ui.DefaultKeyMap()

	// Test that all keys are properly initialized
	if len(keys.Reset.Keys()) == 0 {
		t.Error("Reset key binding not initialized")
	}
	if len(keys.Pause.Keys()) == 0 {
		t.Error("Pause key binding not initialized")
	}
	if len(keys.Stats.Keys()) == 0 {
		t.Error("Stats key binding not initialized")
	}
	if len(keys.Quit.Keys()) == 0 {
		t.Error("Quit key binding not initialized")
	}
}
