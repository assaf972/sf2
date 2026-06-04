package fx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// S20-T01: a disabled flanger must pass the dry signal through unchanged.
func TestFlangerBypassIsIdentity(t *testing.T) {
	f := NewFlanger(1000)
	f.SetParam("rate", 0.5)
	f.SetParam("depth", 80)
	f.SetParam("mix", 100)
	// not enabled
	x := []float32{1, 0.5, -0.25, 0, 0, 0, 0, 0}
	want := append([]float32(nil), x...)
	f.Process(x)
	assert.Equal(t, want, x, "disabled flanger must pass the dry signal unchanged")
}

// S20-T02: with the LFO held still (rate 0) and no feedback, the flanger is a
// fixed short delay: an impulse reappears at the 1 ms base-delay offset scaled
// by mix. At sr=1000, 1 ms == 1 sample.
func TestFlangerStaticCombEchoesImpulse(t *testing.T) {
	f := NewFlanger(1000)
	f.SetParam("rate", 0)  // LFO frozen at phase 0 => delay == base (1 ms)
	f.SetParam("depth", 0) // no sweep
	f.SetParam("feedback", 0)
	f.SetParam("mix", 100) // 1.0
	f.SetEnabled(true)

	x := make([]float32, 8)
	x[0] = 1.0 // impulse
	f.Process(x)

	assert.InDelta(t, 1.0, x[0], 1e-6, "dry impulse passes through")
	assert.InDelta(t, 1.0, x[1], 1e-6, "wet copy one sample (1 ms) later at full mix")
	assert.InDelta(t, 0.0, x[2], 1e-6, "no feedback => no further repeats")
}
