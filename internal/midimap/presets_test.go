package midimap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S10-T02: every bundled preset is well-formed — it maps the GM volume/pan CCs
// and every CC number is in the valid 0..127 range.
func TestPresetsAreWellFormed(t *testing.T) {
	for _, name := range Names() {
		m := Get(name)
		require.Equal(t, name, m.Name, "Get(%q) returned wrong map", name)

		// Volume and pan are present and on the GM standards.
		volCC := findCC(m, Volume)
		panCC := findCC(m, Pan)
		require.NotEqual(t, -1, volCC, "%s: missing Volume mapping", name)
		require.NotEqual(t, -1, panCC, "%s: missing Pan mapping", name)
		assert.Equal(t, 7, volCC, "%s: volume should be CC7", name)
		assert.Equal(t, 10, panCC, "%s: pan should be CC10", name)

		for cc, action := range m.CC {
			assert.GreaterOrEqual(t, cc, 0, "%s: CC %d out of range", name, cc)
			assert.LessOrEqual(t, cc, 127, "%s: CC %d out of range", name, cc)
			assert.NotEqual(t, None, action, "%s: CC %d maps to None", name, cc)
		}
	}
}

func TestUnknownPresetFallsBackToGeneric(t *testing.T) {
	assert.Equal(t, "Generic GM", Get("nope").Name)
}

// S23-T01: the GigSynth hardware factory presets exist and wire the onboard
// surface to the full effect chain, including the new Phaser and Flanger.
func TestProductPresetsCoverFullFXChain(t *testing.T) {
	for _, name := range []string{"GigSynth GS-49", "GigSynth GS-61", "GigSynth GS-Desktop"} {
		m := Get(name)
		require.Equal(t, name, m.Name)
		assert.Equal(t, 7, findCC(m, Volume), "%s: volume on CC7", name)
		assert.Equal(t, 10, findCC(m, Pan), "%s: pan on CC10", name)
		// Every product surface reaches a phaser and a flanger parameter.
		assert.NotEqual(t, -1, anyCC(m, PhaserRate, PhaserDepth, PhaserFeedback), "%s: no phaser control", name)
		assert.NotEqual(t, -1, anyCC(m, FlangerRate, FlangerDepth, FlangerFeedback, FlangerMix), "%s: no flanger control", name)
	}
	// GS-61 has the largest surface (8 knobs + 8 sliders) — the most bindings.
	assert.Greater(t, len(Get("GigSynth GS-61").CC), len(Get("GigSynth GS-49").CC))
}

func anyCC(m Map, actions ...Action) int {
	for _, a := range actions {
		if cc := findCC(m, a); cc != -1 {
			return cc
		}
	}
	return -1
}

func findCC(m Map, a Action) int {
	for cc, act := range m.CC {
		if act == a {
			return cc
		}
	}
	return -1
}
