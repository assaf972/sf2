// Package midimap maps a controller's incoming MIDI CC numbers to GigSynth
// mixer/effect actions, so a keyboard's own faders and knobs drive volume, pan
// and the per-keyboard chorus/delay in real time. Each supported manufacturer
// ships a preset; users can also build a custom map.
package midimap

// Action is what a CC controls in GigSynth.
type Action int

const (
	None Action = iota
	Volume
	Pan
	ChorusRate
	ChorusDepth
	DelayTime
	DelayFeedback
	DelayMix
	Sustain
)

// Map binds CC numbers to actions for one keyboard.
type Map struct {
	Name string
	CC   map[int]Action
}

// Lookup returns the action bound to a CC, if any.
func (m Map) Lookup(cc int) (Action, bool) {
	a, ok := m.CC[cc]
	if !ok || a == None {
		return None, false
	}
	return a, true
}

// presets are the bundled manufacturer maps. CC7=volume and CC10=pan are the GM
// standards shared by all; the encoder banks differ per vendor.
var presets = map[string]Map{
	"Generic GM": {Name: "Generic GM", CC: map[int]Action{
		7: Volume, 10: Pan, 64: Sustain,
	}},
	"M-Audio": {Name: "M-Audio", CC: map[int]Action{
		7: Volume, 10: Pan, 64: Sustain,
		74: ChorusDepth, // Enc 1
		71: DelayMix,    // Enc 2
		72: DelayTime,   // Enc 3
		73: ChorusRate,  // Enc 4
	}},
	"Behringer": {Name: "Behringer", CC: map[int]Action{
		7: Volume, 10: Pan, 64: Sustain,
		70: ChorusDepth, 71: DelayMix, 72: DelayTime, 73: DelayFeedback,
	}},
	"Arturia": {Name: "Arturia", CC: map[int]Action{
		7: Volume, 10: Pan, 64: Sustain,
		74: ChorusRate, 71: ChorusDepth, 76: DelayTime, 77: DelayMix,
	}},
	"Novation": {Name: "Novation", CC: map[int]Action{
		7: Volume, 10: Pan, 64: Sustain,
		21: ChorusDepth, 22: DelayMix, 23: DelayTime, 24: DelayFeedback,
	}},
	"Akai": {Name: "Akai", CC: map[int]Action{
		7: Volume, 10: Pan, 64: Sustain,
		20: ChorusDepth, 21: DelayMix, 22: DelayTime, 23: DelayFeedback,
	}},
}

// Names returns the available preset names (for the Settings dropdown).
func Names() []string {
	return []string{"Generic GM", "M-Audio", "Behringer", "Arturia", "Novation", "Akai"}
}

// Presets returns a copy-friendly view of all bundled maps.
func Presets() map[string]Map { return presets }

// Get returns the named preset, or Generic GM if unknown.
func Get(name string) Map {
	if m, ok := presets[name]; ok {
		return m
	}
	return presets["Generic GM"]
}
