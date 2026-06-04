package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S06-T01 (UI side): the Songs view lists songs from the library and refreshes
// when new ones are added.
func TestSongsViewListsSongs(t *testing.T) {
	u, lib := newTestUI(t)
	_, err := lib.DB().CreateSong("Shine On You Crazy Diamond")
	require.NoError(t, err)
	_, err = lib.DB().CreateSong("Jump")
	require.NoError(t, err)

	u.buildSongsView() // builds the list and calls refreshSongs

	require.Len(t, u.songs, 2)
	names := []string{u.songs[0].Name, u.songs[1].Name}
	assert.Equal(t, []string{"Jump", "Shine On You Crazy Diamond"}, names) // ordered by name

	// Adding a song and refreshing updates the bound slice.
	_, err = lib.DB().CreateSong("Aaa First")
	require.NoError(t, err)
	u.refreshSongs()
	require.Len(t, u.songs, 3)
	assert.Equal(t, "Aaa First", u.songs[0].Name)
}
