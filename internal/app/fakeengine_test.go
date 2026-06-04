package app

import (
	"fmt"
	"sync"
	"testing"

	"gigsynth/internal/engine"

	"github.com/stretchr/testify/assert"
)

// S01-T01: the real engine satisfies Synth (compile-time, see engine/synth.go),
// and the FakeEngine records calls in order so tests can assert behaviour.
func TestFakeEngineRecordsCallsInOrder(t *testing.T) {
	var _ engine.Synth = (*FakeEngine)(nil) // FakeEngine is a valid Synth

	f := &FakeEngine{}
	f.NoteOn(0, 60, 100)
	f.SetChannelVolume(0, 90)
	f.NoteOff(0, 60)

	assert.Equal(t, []string{
		"NoteOn ch=0 key=60 vel=100",
		"CC ch=0 ctrl=7 val=90",
		"NoteOff ch=0 key=60",
	}, f.Calls)
	assert.Len(t, f.CallsOf("NoteOn"), 1)
}

// FakeEngine is a test double for engine.Synth. It records every call as a
// human-readable string in order, so tests can assert exactly what the
// Controller asked the synth to do — with no audio device or FluidSynth.
type FakeEngine struct {
	mu      sync.Mutex
	Calls   []string
	presets []engine.Preset
	sfErr   error
}

var _ engine.Synth = (*FakeEngine)(nil)

func (f *FakeEngine) rec(format string, args ...any) {
	f.mu.Lock()
	f.Calls = append(f.Calls, fmt.Sprintf(format, args...))
	f.mu.Unlock()
}

// CallsOf returns the recorded calls whose prefix matches name (e.g. "NoteOn").
func (f *FakeEngine) CallsOf(name string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, c := range f.Calls {
		if len(c) >= len(name) && c[:len(name)] == name {
			out = append(out, c)
		}
	}
	return out
}

// Reset clears the recorded call log.
func (f *FakeEngine) Reset() { f.mu.Lock(); f.Calls = nil; f.mu.Unlock() }

func (f *FakeEngine) LoadSoundFont(path string) ([]engine.Preset, error) {
	f.rec("LoadSoundFont %s", path)
	return f.presets, f.sfErr
}
func (f *FakeEngine) SelectProgram(ch, bank, prog int) error {
	f.rec("SelectProgram ch=%d bank=%d prog=%d", ch, bank, prog)
	return nil
}
func (f *FakeEngine) NoteOn(ch, key, vel int)   { f.rec("NoteOn ch=%d key=%d vel=%d", ch, key, vel) }
func (f *FakeEngine) NoteOff(ch, key int)       { f.rec("NoteOff ch=%d key=%d", ch, key) }
func (f *FakeEngine) CC(ch, ctrl, val int)      { f.rec("CC ch=%d ctrl=%d val=%d", ch, ctrl, val) }
func (f *FakeEngine) PitchBend(ch, value int)   { f.rec("PitchBend ch=%d val=%d", ch, value) }
func (f *FakeEngine) AllNotesOff(ch int)        { f.rec("AllNotesOff ch=%d", ch) }
func (f *FakeEngine) SetChannelVolume(ch, v int){ f.rec("CC ch=%d ctrl=7 val=%d", ch, v) }
func (f *FakeEngine) SetChannelPan(ch, p int)   { f.rec("CC ch=%d ctrl=10 val=%d", ch, p) }
func (f *FakeEngine) SetMasterGain(g float64)   { f.rec("SetMasterGain %.2f", g) }
func (f *FakeEngine) Panic()                    { f.rec("Panic") }
