package app

import (
	"path/filepath"
	"testing"

	"gigsynth/internal/db"

	"github.com/stretchr/testify/require"
)

// S06-T01 (data side): song create/rename/delete persist and list correctly.
func TestLibrarySongCRUD(t *testing.T) {
	d, err := db.Open(filepath.Join(t.TempDir(), "lib.db"))
	require.NoError(t, err)
	lib := NewLibrary(d)

	a, err := lib.DB().CreateSong("Jump")
	require.NoError(t, err)
	_, err = lib.DB().CreateSong("Comfortably Numb")
	require.NoError(t, err)

	songs, err := lib.DB().ListSongs()
	require.NoError(t, err)
	require.Len(t, songs, 2)
	require.Equal(t, "Comfortably Numb", songs[0].Name) // ordered by name
	require.Equal(t, "Jump", songs[1].Name)

	require.NoError(t, lib.DB().RenameSong(a.ID, "Jump (1984)"))
	require.NoError(t, lib.DB().DeleteSong(songs[0].ID))

	songs, err = lib.DB().ListSongs()
	require.NoError(t, err)
	require.Len(t, songs, 1)
	require.Equal(t, "Jump (1984)", songs[0].Name)
}
