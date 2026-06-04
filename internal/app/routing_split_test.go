package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// helper: build a controller with a single enabled layer on channel 0.
func ctrlWithLayer(l Layer) (*Controller, *FakeEngine) {
	f := &FakeEngine{}
	c := NewController(f, nil)
	l.Channel = 0
	l.Enabled = true
	c.SetScene(Scene{Name: "t", MasterGain: 0.6, Layers: []Layer{l}})
	f.Reset()
	return c, f
}

// S17-T01: note routing honours key split, transpose and source, and never
// leaves a stuck note.
func TestRoutingSplitTransposeSource(t *testing.T) {
	t.Run("key split gates note-on", func(t *testing.T) {
		c, f := ctrlWithLayer(Layer{Source: SourceAny, KeyLow: 36, KeyHigh: 59})
		c.routeNote(true, "Virtual Keyboard", 60, 100) // above split -> ignored
		c.routeNote(true, "Virtual Keyboard", 48, 100) // inside split -> sounds
		assert.Equal(t, []string{"NoteOn ch=0 key=48 vel=100"}, f.CallsOf("NoteOn"))
	})

	t.Run("transpose shifts the played note", func(t *testing.T) {
		c, f := ctrlWithLayer(Layer{Source: SourceAny, KeyLow: 0, KeyHigh: 127, Transpose: 12})
		c.routeNote(true, "Virtual Keyboard", 48, 100)
		assert.Equal(t, []string{"NoteOn ch=0 key=60 vel=100"}, f.CallsOf("NoteOn"))
	})

	t.Run("transpose out of MIDI range is dropped", func(t *testing.T) {
		c, f := ctrlWithLayer(Layer{Source: SourceAny, KeyLow: 0, KeyHigh: 127, Transpose: 24})
		c.routeNote(true, "Virtual Keyboard", 120, 100) // 144 > 127 -> dropped
		assert.Empty(t, f.CallsOf("NoteOn"))
	})

	t.Run("source gating", func(t *testing.T) {
		c, f := ctrlWithLayer(Layer{Source: "Device B", KeyLow: 0, KeyHigh: 127})
		c.routeNote(true, "Device A", 60, 100) // wrong device -> ignored
		assert.Empty(t, f.CallsOf("NoteOn"))
		c.routeNote(true, "Device B", 60, 100) // matching device -> sounds
		assert.Len(t, f.CallsOf("NoteOn"), 1)
	})

	t.Run("note-off releases the transposed note (no stuck note)", func(t *testing.T) {
		c, f := ctrlWithLayer(Layer{Source: SourceAny, KeyLow: 0, KeyHigh: 127, Transpose: 12})
		c.routeNote(true, "Virtual Keyboard", 48, 100)
		c.routeNote(false, "Virtual Keyboard", 48, 0)
		assert.Equal(t, []string{"NoteOn ch=0 key=60 vel=100"}, f.CallsOf("NoteOn"))
		assert.Equal(t, []string{"NoteOff ch=0 key=60"}, f.CallsOf("NoteOff"))
	})

	t.Run("changing transpose flushes held notes via AllNotesOff", func(t *testing.T) {
		c, f := ctrlWithLayer(Layer{Source: SourceAny, KeyLow: 0, KeyHigh: 127})
		c.routeNote(true, "Virtual Keyboard", 60, 100)
		f.Reset()
		c.SetLayerTranspose(0, 12)
		assert.Equal(t, []string{"AllNotesOff ch=0"}, f.CallsOf("AllNotesOff"))
	})
}
