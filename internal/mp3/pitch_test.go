package mp3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// S11-T02: pitch in semitones maps to a resample ratio of 2^(n/12).
func TestPitchToRatio(t *testing.T) {
	cases := map[int]float64{
		-2: 0.8909,
		-1: 0.9439,
		0:  1.0000,
		1:  1.0595,
		2:  1.1225,
	}
	for semis, want := range cases {
		got := ratioFor(semis)
		assert.InDelta(t, want, got, 0.001, "ratio for %+d semitones", semis)
	}
	// Sanity: an octave up doubles the rate.
	assert.InDelta(t, 2.0, ratioFor(12), 1e-9)
}
