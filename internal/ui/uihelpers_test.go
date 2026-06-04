package ui

import (
	"fmt"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"

	"gigsynth/internal/app"
	"gigsynth/internal/db"
	"gigsynth/internal/engine"
)

// noopSynth is an engine.Synth that does nothing — lets UI tests drive a real
// Controller with no audio device.
type noopSynth struct{}

func (noopSynth) LoadSoundFont(string) ([]engine.Preset, error) { return nil, nil }
func (noopSynth) SelectProgram(int, int, int) error             { return nil }
func (noopSynth) NoteOn(int, int, int)                          {}
func (noopSynth) NoteOff(int, int)                              {}
func (noopSynth) CC(int, int, int)                              {}
func (noopSynth) PitchBend(int, int)                            {}
func (noopSynth) AllNotesOff(int)                               {}
func (noopSynth) SetChannelVolume(int, int)                     {}
func (noopSynth) SetChannelPan(int, int)                        {}
func (noopSynth) SetMasterGain(float64)                         {}
func (noopSynth) Panic()                                        {}

// newTestUI builds a UI wired to a real Controller + Library, with no audio and
// no MIDI hardware, for headless Fyne tests.
func newTestUI(t *testing.T) (*UI, *app.Library) {
	t.Helper()
	test.NewApp()
	d, err := db.Open(filepath.Join(t.TempDir(), "ui.db"))
	if err != nil {
		t.Fatal(err)
	}
	lib := app.NewLibrary(d)
	store, err := app.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctrl := app.NewController(noopSynth{}, nil)
	ctrl.SetLibrary(lib)

	u := &UI{ctrl: ctrl, midi: nil, store: store, lib: lib, presetByLabel: map[string]engine.Preset{}}
	for n := 0; n <= 127; n++ {
		u.noteOptions = append(u.noteOptions, app.NoteName(n))
	}
	for tt := -24; tt <= 24; tt++ {
		u.transposeOptions = append(u.transposeOptions, fmt.Sprintf("%+d", tt))
	}
	return u, lib
}
