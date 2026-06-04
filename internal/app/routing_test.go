package app

import (
	"testing"

	"gigsynth/internal/midiio"

	"github.com/stretchr/testify/assert"
)

// S09-T01: with three keyboards bound to different layers, a note from one
// device only triggers layers whose source is that device or "All keyboards".
func TestPerDeviceRoutingIsolatesKeyboards(t *testing.T) {
	f := &FakeEngine{}
	c := NewController(f, nil)
	c.SetScene(Scene{
		Name:       "t",
		MasterGain: 0.6,
		Layers: []Layer{
			{Channel: 0, Enabled: true, Volume: 100, Pan: 64, Source: "Device A", KeyLow: 0, KeyHigh: 127},
			{Channel: 1, Enabled: true, Volume: 100, Pan: 64, Source: "Device B", KeyLow: 0, KeyHigh: 127},
			{Channel: 2, Enabled: true, Volume: 100, Pan: 64, Source: SourceAny, KeyLow: 0, KeyHigh: 127},
		},
	})
	f.Reset()

	// A note from Device B (via the normalized MIDI event path).
	c.handleMIDI(midiio.Event{Device: "Device B", Type: midiio.NoteOn, Key: 64, Velocity: 100})

	on := f.CallsOf("NoteOn")
	assert.ElementsMatch(t, []string{
		"NoteOn ch=1 key=64 vel=100", // Device B layer
		"NoteOn ch=2 key=64 vel=100", // All-keyboards layer
	}, on, "Device A layer (ch0) must NOT sound for a Device B note")
}
