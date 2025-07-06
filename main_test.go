package main

import (
	"testing"
	"time"
)

func TestBandwidthMonitor(t *testing.T) {
	monitor := NewBandwidthMonitor()
	if monitor == nil {
		t.Fatal("Failed to create bandwidth monitor")
	}
	
	// Test getting current rates
	upload, download, err := monitor.GetCurrentRates()
	if err != nil {
		t.Logf("Warning: Could not get bandwidth rates: %v", err)
	}
	
	t.Logf("Current rates - Upload: %d B/s, Download: %d B/s", upload, download)
}

func TestBrailleChart(t *testing.T) {
	chart := NewBrailleChart(10)
	if chart == nil {
		t.Fatal("Failed to create braille chart")
	}
	
	// Test adding data points
	chart.AddDataPoint(1024, 2048)
	chart.AddDataPoint(2048, 1024)
	
	// Test rendering
	rendered := chart.Render()
	if rendered == "" {
		t.Fatal("Chart rendered empty string")
	}
	
	t.Logf("Chart rendered successfully: %d characters", len(rendered))
}

func TestFormatBandwidth(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B/s"},
		{1024, "1.00 KB/s"},
		{1024 * 1024, "1.00 MB/s"},
		{1024 * 1024 * 1024, "1.00 GB/s"},
	}
	
	for _, test := range tests {
		result := formatBandwidth(test.input)
		if result != test.expected {
			t.Errorf("formatBandwidth(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestUIComponents(t *testing.T) {
	ui := NewUIComponents()
	if ui == nil {
		t.Fatal("Failed to create UI components")
	}
	
	// Test stats rendering
	stats := ui.RenderStats(1024, 2048)
	if stats == "" {
		t.Fatal("Stats rendered empty string")
	}
	
	t.Logf("Stats rendered successfully: %d characters", len(stats))
}

func TestTickMsg(t *testing.T) {
	now := time.Now()
	msg := tickMsg(now)
	
	if time.Time(msg) != now {
		t.Error("tickMsg conversion failed")
	}
}

func BenchmarkBandwidthMonitor(b *testing.B) {
	monitor := NewBandwidthMonitor()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetCurrentRates()
	}
}

func BenchmarkBrailleChart(b *testing.B) {
	chart := NewBrailleChart(120)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chart.AddDataPoint(uint64(i*1024), uint64(i*2048))
		chart.Render()
	}
}
