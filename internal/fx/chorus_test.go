package fx

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// S12-T02: disabled chorus is identity; enabled chorus changes the signal but
// keeps it bounded.
func TestChorusBypassAndBounded(t *testing.T) {
	// Bypass identity.
	c := NewChorus(1000)
	x := []float32{0.3, -0.2, 0.9, -0.7, 0.1}
	want := append([]float32(nil), x...)
	c.Process(x)
	assert.Equal(t, want, x, "disabled chorus must be sample-for-sample identity")

	// Enabled: process a sine; output must differ somewhere and stay bounded.
	c2 := NewChorus(1000)
	c2.SetParam("rate", 2)
	c2.SetParam("depth", 80)
	c2.SetEnabled(true)
	n := 200
	in := make([]float32, n)
	for i := range in {
		in[i] = float32(math.Sin(2 * math.Pi * 5 * float64(i) / 1000))
	}
	out := append([]float32(nil), in...)
	c2.Process(out)

	differs := false
	for i := range out {
		assert.LessOrEqual(t, math.Abs(float64(out[i])), 1.5, "chorus output must stay bounded")
		if math.Abs(float64(out[i]-in[i])) > 1e-4 {
			differs = true
		}
	}
	assert.True(t, differs, "enabled chorus must change the signal")
}
