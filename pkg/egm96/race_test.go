package egm96

import (
	"sync"
	"testing"
)

// TestConcurrentHeightAboveMSL drives HeightAboveMSL from many goroutines
// simultaneously to exercise the lazy grid-load path under contention.
//
// Pre-fix (issue #4) this test failed under `go test -race`: two goroutines
// would both pass the `if len(egm96Grid)==0` check and race on the slice
// assignment in loadEGM96Grid. Post-fix the load is sync.Once-protected and
// this test must pass cleanly.
func TestConcurrentHeightAboveMSL(t *testing.T) {
	const goroutines = 64
	var wg sync.WaitGroup
	wg.Add(goroutines)
	start := make(chan struct{})
	for i := range goroutines {
		go func(seed int) {
			defer wg.Done()
			<-start
			lat := float64(seed%170 - 85) // -85 .. 84
			lng := float64((seed * 7) % 360) // 0 .. 359 (HeightAboveMSL expects 0..360)
			loc := NewLocationGeodetic(lat, lng, 0)
			if _, err := loc.HeightAboveMSL(); err != nil {
				t.Errorf("HeightAboveMSL(%v, %v): %v", lat, lng, err)
			}
		}(i)
	}
	close(start) // release all goroutines together
	wg.Wait()
}
