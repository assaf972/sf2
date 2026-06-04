package fx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// S12-T01: an impulse into the delay reappears after `time` samples, scaled by
// mix, with feedback echoes decaying by the feedback amount.
func TestDelayProducesDecayingEchoes(t *testing.T) {
	d := NewDelay(1000) // sr=1000 => 1 ms == 1 sample
	d.SetParam("time", 4)
	d.SetParam("feedback", 50) // 0.5
	d.SetParam("mix", 100)     // 1.0
	d.SetEnabled(true)

	x := make([]float32, 10)
	x[0] = 1.0 // impulse
	d.Process(x)

	// Dry impulse stays at 0; first echo at sample 4, second (decayed) at 8.
	assert.InDelta(t, 1.0, x[0], 1e-6)
	assert.InDelta(t, 1.0, x[4], 1e-6, "first echo at the delay time")
	assert.InDelta(t, 0.5, x[8], 1e-6, "second echo decayed by feedback 0.5")
	assert.InDelta(t, 0.0, x[5], 1e-6)
}

func TestDelayBypassIsIdentity(t *testing.T) {
	d := NewDelay(1000)
	d.SetParam("time", 4)
	d.SetParam("mix", 100)
	// not enabled
	x := []float32{1, 0.5, -0.25, 0, 0, 0, 0, 0}
	want := append([]float32(nil), x...)
	d.Process(x)
	assert.Equal(t, want, x, "disabled delay must pass the dry signal unchanged")
}
