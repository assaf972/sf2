package app

import (
	"path/filepath"
	"testing"

	"gigsynth/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S07-T01: a part edited and saved is recalled identically (sound, mix, split,
// transpose, pan and effects), and recalling it drives the engine.
func TestPartEditorSaveAndRecall(t *testing.T) {
	d, err := db.Open(filepath.Join(t.TempDir(), "ed.db"))
	require.NoError(t, err)
	lib := NewLibrary(d)
	song, _ := d.CreateSong("S")
	part, _ := d.AddPart(song.ID, "P")

	// Edit a scene: layer 0 with a sound, mix, split, transpose and full FX.
	sc := DefaultScene()
	l := &sc.Layers[0]
	l.PresetName, l.Bank, l.Program = "Fender Rhodes", 1, 5
	l.Volume, l.Pan, l.Enabled = 108, 70, true
	l.KeyLow, l.KeyHigh, l.Transpose = 36, 96, 12
	l.ChorusOn, l.ChorusRate, l.ChorusDepth = true, 0.8, 55
	l.DelayOn, l.DelayTime, l.DelayFeedback, l.DelayMix = true, 320, 35, 30

	require.NoError(t, lib.SaveScene(part.ID, "P", sc))

	// Recall and assert every field round-tripped.
	got, err := lib.SceneForPart(part.ID)
	require.NoError(t, err)
	g := got.Layers[0]
	assert.Equal(t, "Fender Rhodes", g.PresetName)
	assert.Equal(t, [4]int{1, 5, 108, 70}, [4]int{g.Bank, g.Program, g.Volume, g.Pan})
	assert.Equal(t, [3]int{36, 96, 12}, [3]int{g.KeyLow, g.KeyHigh, g.Transpose})
	assert.True(t, g.ChorusOn)
	assert.Equal(t, 0.8, g.ChorusRate)
	assert.Equal(t, 55, g.ChorusDepth)
	assert.True(t, g.DelayOn)
	assert.Equal(t, [3]int{320, 35, 30}, [3]int{g.DelayTime, g.DelayFeedback, g.DelayMix})

	// Recalling through a controller (with a SoundFont loaded) drives the engine.
	f := &FakeEngine{}
	c := NewController(f, nil)
	require.NoError(t, c.LoadSoundFont("x.sf2"))
	f.Reset()
	c.SetScene(got)
	assert.Contains(t, f.Calls, "SelectProgram ch=0 bank=1 prog=5")
	assert.Contains(t, f.Calls, "CC ch=0 ctrl=7 val=108")
	assert.Contains(t, f.Calls, "CC ch=0 ctrl=10 val=70")
}
