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
	PhaserRate
	PhaserDepth
	PhaserFeedback
	FlangerRate
	FlangerDepth
	FlangerFeedback
	FlangerMix
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

	// Factory presets for the GigSynth hardware product line. The onboard knobs
	// and sliders are wired so the control surface fully drives the app for the
	// keyboard it's bound to. Volume is CC7 and pan CC10 (GM standards); the rest
	// of the surface covers the four-effect chain.
	"GigSynth GS-49": {Name: "GigSynth GS-49", CC: map[int]Action{
		// 4 sliders
		7: Volume, 11: ChorusDepth, 12: DelayMix, 13: DelayTime,
		// 4 knobs
		10: Pan, 20: PhaserDepth, 21: FlangerMix, 22: PhaserRate,
		64: Sustain,
	}},
	"GigSynth GS-61": {Name: "GigSynth GS-61", CC: map[int]Action{
		// 8 sliders
		7: Volume, 11: ChorusRate, 12: ChorusDepth, 13: DelayTime,
		14: DelayFeedback, 15: DelayMix, 16: PhaserRate, 17: PhaserDepth,
		// 8 knobs
		10: Pan, 20: PhaserFeedback, 21: FlangerRate, 22: FlangerDepth,
		23: FlangerFeedback, 24: FlangerMix, 25: ChorusRate, 26: DelayMix,
		64: Sustain,
	}},
	"GigSynth GS-Desktop": {Name: "GigSynth GS-Desktop", CC: map[int]Action{
		// 4 sliders
		7: Volume, 11: ChorusDepth, 12: DelayMix, 13: PhaserDepth,
		// 4 knobs
		10: Pan, 20: FlangerMix, 21: DelayTime, 22: ChorusRate,
		64: Sustain,
	}},
}

// Names returns the available preset names (for the Settings dropdown).
func Names() []string {
	return []string{"Generic GM", "M-Audio", "Behringer", "Arturia", "Novation", "Akai",
		"GigSynth GS-49", "GigSynth GS-61", "GigSynth GS-Desktop"}
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
