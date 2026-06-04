// Package engine is a thin, concurrency-safe Go wrapper around libfluidsynth.
//
// FluidSynth runs synthesis on its own real-time audio thread (created by
// new_fluid_audio_driver). The Go side only issues control/MIDI calls into the
// C API, which are cheap and non-blocking, so Go's garbage collector never
// touches the audio hot path. That is what makes Go a good fit here.
package engine

/*
#cgo pkg-config: fluidsynth
#include <fluidsynth.h>
#include <stdlib.h>

// Helper: iterate presets of a loaded soundfont into caller-provided callback
// is awkward across cgo, so we expose small accessors instead and iterate in Go.
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

// Preset describes one instrument program inside a loaded SoundFont.
type Preset struct {
	SFontID int
	Bank    int
	Program int
	Name    string
}

// Engine owns a single FluidSynth instance (settings, synth, audio driver).
type Engine struct {
	mu       sync.Mutex
	settings *C.fluid_settings_t
	synth    *C.fluid_synth_t
	driver   *C.fluid_audio_driver_t
	sfontID  C.int
	loaded   bool
}

// Config controls how the audio engine is created.
type Config struct {
	// AudioDriver is the fluidsynth audio driver name. Empty = let fluidsynth
	// pick the platform default (coreaudio / alsa-or-pipewire / wasapi).
	AudioDriver string
	SampleRate  float64 // e.g. 48000
	// PeriodSize and Periods control latency. Smaller = lower latency but more
	// CPU and higher xrun risk. 64 x 2 @ 48k is a good low-latency live default.
	PeriodSize int
	Periods    int
	// PolyPhony is the max simultaneous voices.
	Polyphony int
}

// DefaultConfig returns sane low-latency defaults for live performance.
func DefaultConfig() Config {
	return Config{
		AudioDriver: "", // platform default
		SampleRate:  48000,
		PeriodSize:  64,
		Periods:     2,
		Polyphony:   256,
	}
}

// New creates and starts a FluidSynth engine.
func New(cfg Config) (*Engine, error) {
	settings := C.new_fluid_settings()
	if settings == nil {
		return nil, fmt.Errorf("engine: new_fluid_settings failed")
	}

	setStr := func(key, val string) {
		ck := C.CString(key)
		cv := C.CString(val)
		C.fluid_settings_setstr(settings, ck, cv)
		C.free(unsafe.Pointer(ck))
		C.free(unsafe.Pointer(cv))
	}
	setInt := func(key string, val int) {
		ck := C.CString(key)
		C.fluid_settings_setint(settings, ck, C.int(val))
		C.free(unsafe.Pointer(ck))
	}
	setNum := func(key string, val float64) {
		ck := C.CString(key)
		C.fluid_settings_setnum(settings, ck, C.double(val))
		C.free(unsafe.Pointer(ck))
	}

	if cfg.AudioDriver != "" {
		setStr("audio.driver", cfg.AudioDriver)
	}
	if cfg.SampleRate > 0 {
		setNum("synth.sample-rate", cfg.SampleRate)
	}
	if cfg.PeriodSize > 0 {
		setInt("audio.period-size", cfg.PeriodSize)
	}
	if cfg.Periods > 0 {
		setInt("audio.periods", cfg.Periods)
	}
	if cfg.Polyphony > 0 {
		setInt("synth.polyphony", cfg.Polyphony)
	}
	// We manage volume per channel via CC7; keep master gain at unity-ish.
	setNum("synth.gain", 0.6)

	synth := C.new_fluid_synth(settings)
	if synth == nil {
		C.delete_fluid_settings(settings)
		return nil, fmt.Errorf("engine: new_fluid_synth failed")
	}

	driver := C.new_fluid_audio_driver(settings, synth)
	if driver == nil {
		C.delete_fluid_synth(synth)
		C.delete_fluid_settings(settings)
		return nil, fmt.Errorf("engine: new_fluid_audio_driver failed (no audio device?)")
	}

	return &Engine{
		settings: settings,
		synth:    synth,
		driver:   driver,
		sfontID:  -1,
	}, nil
}

// Close stops audio and frees all C resources.
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.driver != nil {
		C.delete_fluid_audio_driver(e.driver)
		e.driver = nil
	}
	if e.synth != nil {
		C.delete_fluid_synth(e.synth)
		e.synth = nil
	}
	if e.settings != nil {
		C.delete_fluid_settings(e.settings)
		e.settings = nil
	}
}

// LoadSoundFont loads an .sf2 file, replacing any previously loaded one, and
// returns the list of presets it contains.
func (e *Engine) LoadSoundFont(path string) ([]Preset, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.synth == nil {
		return nil, fmt.Errorf("engine: closed")
	}

	if e.loaded && e.sfontID >= 0 {
		C.fluid_synth_sfunload(e.synth, e.sfontID, 1)
		e.loaded = false
		e.sfontID = -1
	}

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	id := C.fluid_synth_sfload(e.synth, cpath, 1)
	if id == C.FLUID_FAILED {
		return nil, fmt.Errorf("engine: failed to load soundfont %q", path)
	}
	e.sfontID = id
	e.loaded = true

	return e.listPresetsLocked(), nil
}

// listPresetsLocked enumerates presets of the currently loaded soundfont.
// Caller must hold e.mu.
func (e *Engine) listPresetsLocked() []Preset {
	var out []Preset
	sfont := C.fluid_synth_get_sfont_by_id(e.synth, e.sfontID)
	if sfont == nil {
		return out
	}
	C.fluid_sfont_iteration_start(sfont)
	for {
		preset := C.fluid_sfont_iteration_next(sfont)
		if preset == nil {
			break
		}
		name := C.GoString(C.fluid_preset_get_name(preset))
		bank := int(C.fluid_preset_get_banknum(preset))
		prog := int(C.fluid_preset_get_num(preset))
		out = append(out, Preset{
			SFontID: int(e.sfontID),
			Bank:    bank,
			Program: prog,
			Name:    name,
		})
	}
	return out
}

// SelectProgram assigns a preset (bank/program) to a MIDI channel.
func (e *Engine) SelectProgram(channel, bank, program int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.synth == nil || !e.loaded {
		return fmt.Errorf("engine: no soundfont loaded")
	}
	rc := C.fluid_synth_program_select(e.synth, C.int(channel), e.sfontID, C.int(bank), C.int(program))
	if rc == C.FLUID_FAILED {
		return fmt.Errorf("engine: program_select ch=%d bank=%d prog=%d failed", channel, bank, program)
	}
	return nil
}

// NoteOn / NoteOff / CC / PitchBend forward MIDI events to a channel.
// These are called from the MIDI input goroutine on every event, so they take
// the lock only briefly; fluidsynth's own calls are real-time safe.

func (e *Engine) NoteOn(channel, key, vel int) {
	e.mu.Lock()
	if e.synth != nil {
		C.fluid_synth_noteon(e.synth, C.int(channel), C.int(key), C.int(vel))
	}
	e.mu.Unlock()
}

func (e *Engine) NoteOff(channel, key int) {
	e.mu.Lock()
	if e.synth != nil {
		C.fluid_synth_noteoff(e.synth, C.int(channel), C.int(key))
	}
	e.mu.Unlock()
}

func (e *Engine) CC(channel, ctrl, val int) {
	e.mu.Lock()
	if e.synth != nil {
		C.fluid_synth_cc(e.synth, C.int(channel), C.int(ctrl), C.int(val))
	}
	e.mu.Unlock()
}

func (e *Engine) PitchBend(channel, value int) {
	e.mu.Lock()
	if e.synth != nil {
		C.fluid_synth_pitch_bend(e.synth, C.int(channel), C.int(value))
	}
	e.mu.Unlock()
}

// AllNotesOff sends an "all notes off" to a channel (releases held notes).
func (e *Engine) AllNotesOff(channel int) {
	e.mu.Lock()
	if e.synth != nil {
		C.fluid_synth_all_notes_off(e.synth, C.int(channel))
	}
	e.mu.Unlock()
}

// Panic silences every channel immediately.
func (e *Engine) Panic() {
	e.mu.Lock()
	if e.synth != nil {
		C.fluid_synth_system_reset(e.synth)
	}
	e.mu.Unlock()
}

// SetMasterGain sets the global output gain (0.0 .. 10.0, but keep <=2 live).
func (e *Engine) SetMasterGain(g float64) {
	e.mu.Lock()
	if e.synth != nil {
		C.fluid_synth_set_gain(e.synth, C.float(g))
	}
	e.mu.Unlock()
}

// SetChannelVolume sets per-channel volume via CC7 (0..127). This is how the
// mixer balances channels when layering a patch.
func (e *Engine) SetChannelVolume(channel, vol int) { e.CC(channel, 7, vol) }

// SetChannelPan sets per-channel pan via CC10 (0=left .. 64=center .. 127=right).
func (e *Engine) SetChannelPan(channel, pan int) { e.CC(channel, 10, pan) }
