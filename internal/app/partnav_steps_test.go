package app

import (
	"fmt"
	"path/filepath"
	"testing"

	"gigsynth/internal/db"

	"github.com/cucumber/godog"
)

func TestPartNavigationFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var (
				c      *Controller
				f      *FakeEngine
				store  *db.DB
				songID int64
			)
			dir := t.TempDir()

			sc.Step(`^a song "([^"]*)" with parts:$`, func(name string, tbl *godog.Table) error {
				f = &FakeEngine{}
				c = NewController(f, nil)
				var err error
				store, err = db.Open(filepath.Join(dir, "nav.db"))
				if err != nil {
					return err
				}
				c.SetLibrary(NewLibrary(store))
				song, err := store.CreateSong(name)
				if err != nil {
					return err
				}
				songID = song.ID
				for _, row := range tbl.Rows[1:] { // skip header
					partName, sound := row.Cells[1].Value, row.Cells[2].Value
					p, err := store.AddPart(songID, partName)
					if err != nil {
						return err
					}
					ch := p.Channels[0]
					ch.InstrumentName, ch.Enabled = sound, true
					if err := store.SetPartChannel(p.ID, ch); err != nil {
						return err
					}
				}
				return nil
			})
			sc.Step(`^part 1 is active$`, func() error { return c.LoadSong(songID, "Shine On You Crazy Diamond") })
			sc.Step(`^part 3 is active$`, func() error { c.GoToPart(2); return nil })
			sc.Step(`^I press "Next Part"$`, func() error { f.Reset(); c.NextPart(); return nil })
			sc.Step(`^I press "Prev Part"$`, func() error { f.Reset(); c.PrevPart(); return nil })
			sc.Step(`^the active part is "([^"]*)"$`, func(want string) error {
				if got := c.CurrentPartName(); got != want {
					return fmt.Errorf("active part = %q, want %q", got, want)
				}
				return nil
			})
			sc.Step(`^keyboard 1 sound is "([^"]*)"$`, func(want string) error {
				if got := c.Scene().Layers[0].PresetName; got != want {
					return fmt.Errorf("keyboard 1 sound = %q, want %q", got, want)
				}
				return nil
			})
			sc.Step(`^a PANIC was issued before the recall$`, func() error {
				if len(f.Calls) == 0 || f.Calls[0] != "Panic" {
					return fmt.Errorf("expected PANIC first on recall, got %v", f.Calls)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "part_navigation.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("part navigation feature has failing scenarios")
	}
}
