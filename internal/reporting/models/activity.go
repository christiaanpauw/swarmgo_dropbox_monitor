package models

import (
	"sort"
	"time"
)

// ActivityPattern represents patterns found in file changes
type ActivityPattern struct {
	TimeRanges     []TimeRange           // Active time ranges
	HotspotDirs    []DirectoryHotspot    // Directories with high activity
	FilePatterns   []FilePattern         // Common file patterns
	ActivityTrends map[string]TrendStats // Activity trends by category
}

// TimeRange represents a period of high activity
type TimeRange struct {
	Start    time.Time
	End      time.Time
	Changes  int
	Patterns []string // Notable patterns during this time
}

// DirectoryHotspot represents a directory with significant activity
type DirectoryHotspot struct {
	Path           string
	ChangeCount    int
	CommonPatterns []string
	LastActive     time.Time
}

// FilePattern represents a common pattern in file changes
type FilePattern struct {
	Pattern     string  // The pattern description
	Occurrences int     // Number of occurrences
	Confidence  float64 // Confidence score (0-1)
}

// TrendStats represents statistical trends in activity
type TrendStats struct {
	Total       int       // Total occurrences
	Peak        time.Time // Time of peak activity
	PeakCount   int       // Count during peak
	Growth      float64   // Growth rate
	Seasonality float64   // Seasonality score (0-1)
}

// NewActivityPattern creates a new activity pattern instance
func NewActivityPattern() *ActivityPattern {
	return &ActivityPattern{
		TimeRanges:     make([]TimeRange, 0),
		HotspotDirs:    make([]DirectoryHotspot, 0),
		FilePatterns:   make([]FilePattern, 0),
		ActivityTrends: make(map[string]TrendStats),
	}
}

// AddTimeRange adds a new time range to the activity pattern
func (ap *ActivityPattern) AddTimeRange(tr TimeRange) {
	ap.TimeRanges = append(ap.TimeRanges, tr)
}

// AddHotspot adds a new directory hotspot
func (ap *ActivityPattern) AddHotspot(hs DirectoryHotspot) {
	ap.HotspotDirs = append(ap.HotspotDirs, hs)
}

// AddFilePattern adds a new file pattern
func (ap *ActivityPattern) AddFilePattern(fp FilePattern) {
	ap.FilePatterns = append(ap.FilePatterns, fp)
}

// UpdateTrend updates activity trends for a category
func (ap *ActivityPattern) UpdateTrend(category string, stats TrendStats) {
	ap.ActivityTrends[category] = stats
}

// GetMostActiveTimeRange returns the time range with the most changes
func (ap *ActivityPattern) GetMostActiveTimeRange() *TimeRange {
	if len(ap.TimeRanges) == 0 {
		return nil
	}

	mostActive := &ap.TimeRanges[0]
	for i := range ap.TimeRanges {
		if ap.TimeRanges[i].Changes > mostActive.Changes {
			mostActive = &ap.TimeRanges[i]
		}
	}
	return mostActive
}

// GetTopHotspots returns the n most active directory hotspots
func (ap *ActivityPattern) GetTopHotspots(n int) []DirectoryHotspot {
	if len(ap.HotspotDirs) <= n {
		return ap.HotspotDirs
	}

	// Sort hotspots by change count
	sorted := make([]DirectoryHotspot, len(ap.HotspotDirs))
	copy(sorted, ap.HotspotDirs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ChangeCount > sorted[j].ChangeCount
	})

	return sorted[:n]
}
