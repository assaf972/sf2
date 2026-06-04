package app

// NumLayers is the number of mixer channels/layers the app exposes. The user
// asked for "two channels or maybe four"; we build four and let the UI/scene
// enable as many as needed.
const NumLayers = 4

// MaxKeyboards is the number of MIDI keyboards we route simultaneously.
const MaxKeyboards = 3

// SourceAny is the routing value meaning "any connected keyboard".
const SourceAny = "All keyboards"

// Layer is one mixer channel: a sound (SF2 preset) plus its mix settings and
// MIDI routing (which keyboard drives it, key split range, transpose).
type Layer struct {
	Channel    int    `json:"channel"`    // fluidsynth MIDI channel (fixed per slot)
	Name       string `json:"name"`       // user label, e.g. "Piano", "Strings"
	Bank       int    `json:"bank"`       // SF2 bank
	Program    int    `json:"program"`    // SF2 program
	PresetName string `json:"presetName"` // display name of the SF2 preset
	Volume     int    `json:"volume"`     // 0..127 (CC7)
	Pan        int    `json:"pan"`        // 0..127 (CC10), 64 = center
	Mute       bool   `json:"mute"`
	Enabled    bool   `json:"enabled"`

	// Routing
	Source    string `json:"source"`    // device name or SourceAny
	KeyLow    int    `json:"keyLow"`    // split low note (0..127)
	KeyHigh   int    `json:"keyHigh"`   // split high note (0..127)
	Transpose int    `json:"transpose"` // semitones, +/-

	// Per-keyboard effects (S13 chorus / S14 delay).
	ChorusOn      bool    `json:"chorusOn"`
	ChorusRate    float64 `json:"chorusRate"`    // Hz
	ChorusDepth   int     `json:"chorusDepth"`   // %
	DelayOn       bool    `json:"delayOn"`
	DelayTime     int     `json:"delayTime"`     // ms
	DelayFeedback int     `json:"delayFeedback"` // %
	DelayMix      int     `json:"delayMix"`      // %
}

// Scene is a recallable performance patch: the full set of layer assignments
// and mix. This is the "preset" the gigging player selects per song.
type Scene struct {
	Name       string  `json:"name"`
	MasterGain float64 `json:"masterGain"`
	Layers     []Layer `json:"layers"`
}

func defaultLayer(slot int) Layer {
	return Layer{
		Channel:   slot,
		Name:      []string{"Layer 1", "Layer 2", "Layer 3", "Layer 4"}[slot],
		Bank:      0,
		Program:   0,
		Volume:    100,
		Pan:       64,
		Mute:      false,
		Enabled:   slot < 2, // first two on by default
		Source:    SourceAny,
		KeyLow:    0,
		KeyHigh:   127,
		Transpose: 0,
		// FX defaults: off, with musically sensible starting values.
		ChorusRate:    0.8,
		ChorusDepth:   50,
		DelayTime:     300,
		DelayFeedback: 30,
		DelayMix:      25,
	}
}

// DefaultScene returns a blank 4-layer scene.
func DefaultScene() Scene {
	layers := make([]Layer, NumLayers)
	for i := 0; i < NumLayers; i++ {
		layers[i] = defaultLayer(i)
	}
	return Scene{Name: "Init", MasterGain: 0.6, Layers: layers}
}
