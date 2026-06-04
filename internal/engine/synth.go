package engine

// Synth is the control surface the application uses to drive a synthesizer.
//
// It is deliberately small and free of any cgo/C types so the domain layer
// (internal/app) can be built and tested against a fake implementation with no
// audio device or FluidSynth present. The real *Engine (cgo FluidSynth) and the
// test FakeEngine both satisfy it.
type Synth interface {
	// SoundFont / program selection
	LoadSoundFont(path string) ([]Preset, error)
	SelectProgram(channel, bank, program int) error

	// Live MIDI
	NoteOn(channel, key, vel int)
	NoteOff(channel, key int)
	CC(channel, ctrl, val int)
	PitchBend(channel, value int)
	AllNotesOff(channel int)

	// Mixer / global
	SetChannelVolume(channel, vol int)
	SetChannelPan(channel, pan int)
	SetMasterGain(g float64)
	Panic()
}

// Compile-time assertion: the real cgo engine implements the interface.
var _ Synth = (*Engine)(nil)
