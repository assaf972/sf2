package app

import (
	"path/filepath"
	"testing"

	"gigsynth/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S08-T01: settings persist to the database and reload on next open; defaults
// apply when nothing has been saved.
func TestSettingsPersistAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.db")

	d1, err := db.Open(path)
	require.NoError(t, err)

	// Defaults when nothing saved yet.
	got, err := LoadSettings(d1)
	require.NoError(t, err)
	assert.Equal(t, DefaultSettings(), got)

	// Save custom settings, then close.
	want := Settings{
		SF2Path: "/opt/gigsynth/Live.sf2", AudioDriver: "coreaudio",
		SampleRate: 48000, PeriodSize: 128, Periods: 3, Polyphony: 256,
		MP3Folder: "/music/backing", Theme: "Dark (stage)", StartTab: "Live",
	}
	require.NoError(t, SaveSettings(d1, want))
	require.NoError(t, d1.Close())

	// Reopen: settings survive the restart.
	d2, err := db.Open(path)
	require.NoError(t, err)
	defer d2.Close()
	got, err = LoadSettings(d2)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
