package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// S01-T02: a note on an enabled layer produces exactly one NoteOn on the fake
// engine at that layer's channel — proving the Controller is fully testable
// without cgo or an audio device.
func TestControllerRoutesNoteToEngine(t *testing.T) {
	f := &FakeEngine{}
	c := NewController(f, nil) // nil MIDI manager: we drive routing directly

	// Default scene has layers 0 and 1 enabled across the whole keyboard.
	c.SetScene(Scene{
		Name:       "t",
		MasterGain: 0.6,
		Layers: []Layer{
			{Channel: 0, Enabled: true, Volume: 100, Pan: 64, Source: SourceAny, KeyLow: 0, KeyHigh: 127},
			{Channel: 1, Enabled: false, Volume: 100, Pan: 64, Source: SourceAny, KeyLow: 0, KeyHigh: 127},
		},
	})
	f.Reset() // ignore the program/CC calls from applying the scene

	c.VirtualNoteOn(60, 100)

	assert.Equal(t, []string{"NoteOn ch=0 key=60 vel=100"}, f.CallsOf("NoteOn"),
		"only the enabled layer (channel 0) should sound")
}
