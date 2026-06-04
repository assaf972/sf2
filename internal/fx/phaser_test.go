package fx

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// S19-T01: a disabled phaser must pass the dry signal through unchanged.
func TestPhaserBypassIsIdentity(t *testing.T) {
	p := NewPhaser(48000)
	p.SetParam("rate", 0.5)
	p.SetParam("depth", 80)
	p.SetParam("feedback", 30)
	// not enabled
	x := []float32{1, 0.5, -0.25, 0, 0.75, -1, 0, 0}
	want := append([]float32(nil), x...)
	p.Process(x)
	assert.Equal(t, want, x, "disabled phaser must pass the dry signal unchanged")
}

// S19-T02: an enabled phaser colours the signal (output differs from dry) and
// stays bounded/finite (the all-pass cascade with feedback must be stable).
func TestPhaserAltersSignalAndStaysStable(t *testing.T) {
	p := NewPhaser(48000)
	p.SetParam("rate", 1.0)
	p.SetParam("depth", 100)
	p.SetParam("feedback", 90)
	p.SetEnabled(true)

	x := make([]float32, 2000)
	for i := range x {
		x[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / 48000)) // 440 Hz tone
	}
	dry := append([]float32(nil), x...)
	p.Process(x)

	differs := false
	for i := range x {
		assert.False(t, math.IsNaN(float64(x[i])) || math.IsInf(float64(x[i]), 0), "output must stay finite")
		assert.LessOrEqual(t, math.Abs(float64(x[i])), 4.0, "output must stay bounded")
		if math.Abs(float64(x[i]-dry[i])) > 1e-4 {
			differs = true
		}
	}
	assert.True(t, differs, "enabled phaser must change the signal")
}
