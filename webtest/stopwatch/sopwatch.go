package stopwatch

import "time"

// Stopwatch is a simple utility to measure elapsed time.
type Stopwatch struct {
	start time.Time
	end   time.Time
}

// Start begins the stopwatch.
func (sw *Stopwatch) Start() {
	sw.start = time.Now()
	sw.end = time.Time{}
}

// Stop ends the stopwatch.
func (sw *Stopwatch) Stop() {
	sw.end = time.Now()
}

// Elapsed returns the elapsed duration.
func (sw *Stopwatch) Elapsed() time.Duration {
	if sw.end.IsZero() {
		return time.Since(sw.start)
	}
	return sw.end.Sub(sw.start)
}
