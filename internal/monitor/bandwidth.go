// Package monitor provides cross-platform bandwidth monitoring functionality
//
// This package provides bandwidth monitoring capabilities using the gopsutil
// library to gather network interface statistics across different platforms.
package monitor

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

// BandwidthMonitor handles cross-platform bandwidth monitoring
type BandwidthMonitor struct {
	lastStats    map[string]net.IOCountersStat
	lastTime     time.Time
	currentRates BandwidthRates
	// Optimization: reuse slice to avoid allocations
	statsBuffer  []net.IOCountersStat
}

// BandwidthRates represents current upload/download rates
type BandwidthRates struct {
	Upload   uint64 // bytes per second
	Download uint64 // bytes per second
}

// NewBandwidthMonitor creates a new bandwidth monitor
func NewBandwidthMonitor() *BandwidthMonitor {
	monitor := &BandwidthMonitor{
		lastStats:   make(map[string]net.IOCountersStat),
		lastTime:    time.Now(),
		statsBuffer: make([]net.IOCountersStat, 0, 10), // Pre-allocate for typical interface count
	}

	// Initialize with first reading
	monitor.updateStats()

	return monitor
}

// GetCurrentRates returns the current upload and download rates
func (bm *BandwidthMonitor) GetCurrentRates() (uint64, uint64, error) {
	err := bm.updateStats()
	if err != nil {
		return 0, 0, err
	}

	return bm.currentRates.Upload, bm.currentRates.Download, nil
}

// updateStats fetches new network statistics and calculates rates
func (bm *BandwidthMonitor) updateStats() error {
	// Get network interface statistics
	stats, err := net.IOCounters(true) // true = per interface
	if err != nil {
		return fmt.Errorf("failed to get network stats: %w", err)
	}

	currentTime := time.Now()
	timeDiff := currentTime.Sub(bm.lastTime).Seconds()

	// Skip if time difference is too small to avoid division by zero
	if timeDiff < 0.01 {
		return nil
	}

	var totalUpload, totalDownload uint64

	// Optimization: calculate rates more efficiently
	timeDiffRecip := 1.0 / timeDiff // Calculate reciprocal once

	// Calculate rates for all interfaces
	for _, stat := range stats {
		// Skip loopback interfaces
		if stat.Name == "lo" || stat.Name == "Loopback" {
			continue
		}

		if lastStat, exists := bm.lastStats[stat.Name]; exists {
			// Calculate bytes transferred since last measurement
			bytesSent := stat.BytesSent - lastStat.BytesSent
			bytesRecv := stat.BytesRecv - lastStat.BytesRecv

			// Handle counter rollover (unlikely with 64-bit counters)
			if stat.BytesSent < lastStat.BytesSent {
				bytesSent = stat.BytesSent
			}
			if stat.BytesRecv < lastStat.BytesRecv {
				bytesRecv = stat.BytesRecv
			}

			// Convert to rate (bytes per second) - use reciprocal for efficiency
			uploadRate := uint64(float64(bytesSent) * timeDiffRecip)
			downloadRate := uint64(float64(bytesRecv) * timeDiffRecip)

			totalUpload += uploadRate
			totalDownload += downloadRate
		}

		// Update last stats
		bm.lastStats[stat.Name] = stat
	}

	// Update current rates
	bm.currentRates.Upload = totalUpload
	bm.currentRates.Download = totalDownload
	bm.lastTime = currentTime

	return nil
}
