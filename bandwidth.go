package main

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
}

// BandwidthRates represents current upload/download rates
type BandwidthRates struct {
	Upload   uint64 // bytes per second
	Download uint64 // bytes per second
}

// NewBandwidthMonitor creates a new bandwidth monitor
func NewBandwidthMonitor() *BandwidthMonitor {
	monitor := &BandwidthMonitor{
		lastStats: make(map[string]net.IOCountersStat),
		lastTime:  time.Now(),
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

			// Convert to rate (bytes per second)
			uploadRate := uint64(float64(bytesSent) / timeDiff)
			downloadRate := uint64(float64(bytesRecv) / timeDiff)

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

// GetPeakRates returns the maximum observed rates (for scaling)
func (bm *BandwidthMonitor) GetPeakRates() (uint64, uint64) {
	// This could be enhanced to track historical peaks
	// For now, return current rates
	return bm.currentRates.Upload, bm.currentRates.Download
}
