package app

import (
	"path/filepath"

	"gigsynth/internal/db"
)

// fxWorld is shared scenario state for the chorus and delay features: a
// Controller (FakeEngine) plus a Library with one song and one part to save into.
type fxWorld struct {
	c      *Controller
	lib    *Library
	partID int64
}

func newFXWorld(dir string) (*fxWorld, error) {
	c := NewController(&FakeEngine{}, nil) // DefaultScene: 4 layers, FX off
	d, err := db.Open(filepath.Join(dir, "fx.db"))
	if err != nil {
		return nil, err
	}
	lib := NewLibrary(d)
	song, err := d.CreateSong("S")
	if err != nil {
		return nil, err
	}
	part, err := d.AddPart(song.ID, "P")
	if err != nil {
		return nil, err
	}
	return &fxWorld{c: c, lib: lib, partID: part.ID}, nil
}

func (w *fxWorld) layer(kb int) Layer { return w.c.Scene().Layers[kb-1] }

// saveAndRecall persists the live scene, wipes it, and recalls it from the DB —
// proving FX state survives a round-trip through the part.
func (w *fxWorld) saveAndRecall() error {
	if err := w.lib.SaveScene(w.partID, "P", w.c.Scene()); err != nil {
		return err
	}
	w.c.SetScene(DefaultScene()) // clear live state
	s, err := w.lib.SceneForPart(w.partID)
	if err != nil {
		return err
	}
	w.c.SetScene(s)
	return nil
}
