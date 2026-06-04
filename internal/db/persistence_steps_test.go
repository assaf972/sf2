package db

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

// persistWorld holds per-scenario state for the persistence feature.
type persistWorld struct {
	dir       string
	db        *DB
	songIDs   map[string]int64
	partIDs   map[string]int64
	curSong   int64
	curPart   int64
	reloaded  Part
}

func (w *persistWorld) openDB() error {
	d, err := Open(filepath.Join(w.dir, "test.db"))
	if err != nil {
		return err
	}
	w.db = d
	w.songIDs = map[string]int64{}
	w.partIDs = map[string]int64{}
	return nil
}

func (w *persistWorld) emptyDB() error { return w.openDB() }

func (w *persistWorld) createSong(name string) error {
	s, err := w.db.CreateSong(name)
	if err != nil {
		return err
	}
	w.songIDs[name] = s.ID
	w.curSong = s.ID
	return nil
}

func (w *persistWorld) addPart(name string) error {
	p, err := w.db.AddPart(w.curSong, name)
	if err != nil {
		return err
	}
	w.partIDs[name] = p.ID
	w.curPart = p.ID
	return nil
}

func (w *persistWorld) songWithPart(song, part string) error {
	if err := w.createSong(song); err != nil {
		return err
	}
	return w.addPart(part)
}

func (w *persistWorld) songWithNParts(song string, n int) error {
	if err := w.createSong(song); err != nil {
		return err
	}
	for i := 1; i <= n; i++ {
		if err := w.addPart(fmt.Sprintf("P%d", i)); err != nil {
			return err
		}
	}
	return nil
}

func (w *persistWorld) partsInOrder(want int, table *godog.Table) error {
	parts, err := w.db.ListParts(w.curSong)
	if err != nil {
		return err
	}
	if len(parts) != want {
		return fmt.Errorf("want %d parts, got %d", want, len(parts))
	}
	for i, row := range table.Rows[1:] { // skip header
		name := row.Cells[1].Value
		if parts[i].Name != name {
			return fmt.Errorf("row %d: want %q, got %q", i+1, name, parts[i].Name)
		}
		if parts[i].SortOrder != i+1 {
			return fmt.Errorf("row %d: want sort_order %d, got %d", i+1, i+1, parts[i].SortOrder)
		}
	}
	return nil
}

func (w *persistWorld) setSound(part string, ch, bank, prog, vol, pan int) error {
	c := defaultChannel(ch)
	c.Bank, c.Program, c.Level, c.Pan = bank, prog, vol, pan
	return w.db.SetPartChannel(w.partIDs[part], c)
}

func (w *persistWorld) setChorus(part string, ch int, rate float64, depth int) error {
	c, err := w.channel(part, ch)
	if err != nil {
		return err
	}
	c.ChorusOn, c.ChorusRate, c.ChorusDepth = true, rate, depth
	return w.db.SetPartChannel(w.partIDs[part], c)
}

func (w *persistWorld) setDelay(part string, ch, time, fb, mix int) error {
	c, err := w.channel(part, ch)
	if err != nil {
		return err
	}
	c.DelayOn, c.DelayTime, c.DelayFeedback, c.DelayMix = true, time, fb, mix
	return w.db.SetPartChannel(w.partIDs[part], c)
}

func (w *persistWorld) channel(part string, ch int) (Channel, error) {
	p, err := w.db.GetPart(w.partIDs[part])
	if err != nil {
		return Channel{}, err
	}
	return p.Channels[ch-1], nil
}

func (w *persistWorld) reloadPart() error {
	p, err := w.db.GetPart(w.curPart)
	if err != nil {
		return err
	}
	w.reloaded = p
	return nil
}

func (w *persistWorld) assertSound(ch, bank, prog, vol int) error {
	c := w.reloaded.Channels[ch-1]
	if c.Bank != bank || c.Program != prog || c.Level != vol {
		return fmt.Errorf("ch%d: got bank=%d prog=%d vol=%d", ch, c.Bank, c.Program, c.Level)
	}
	return nil
}

func (w *persistWorld) assertChorus(ch int, rate float64, depth int) error {
	c := w.reloaded.Channels[ch-1]
	if !c.ChorusOn || c.ChorusRate != rate || c.ChorusDepth != depth {
		return fmt.Errorf("ch%d chorus: on=%v rate=%v depth=%d", ch, c.ChorusOn, c.ChorusRate, c.ChorusDepth)
	}
	return nil
}

func (w *persistWorld) assertDelay(ch, time, fb, mix int) error {
	c := w.reloaded.Channels[ch-1]
	if !c.DelayOn || c.DelayTime != time || c.DelayFeedback != fb || c.DelayMix != mix {
		return fmt.Errorf("ch%d delay: on=%v time=%d fb=%d mix=%d", ch, c.DelayOn, c.DelayTime, c.DelayFeedback, c.DelayMix)
	}
	return nil
}

func (w *persistWorld) deleteSong(name string) error { return w.db.DeleteSong(w.songIDs[name]) }

func (w *persistWorld) noParts(song string) error {
	n, err := w.db.CountParts(w.songIDs[song])
	if err != nil {
		return err
	}
	if n != 0 {
		return fmt.Errorf("expected 0 parts after delete, got %d", n)
	}
	return nil
}

func TestPersistenceFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			w := &persistWorld{dir: t.TempDir()}
			sc.Step(`^an empty in-memory GigSynth database$`, w.emptyDB)
			sc.Step(`^I create a song "([^"]*)"$`, w.createSong)
			sc.Step(`^I add a part "([^"]*)" to that song$`, w.addPart)
			sc.Step(`^the song has (\d+) parts in order:$`, w.partsInOrder)
			sc.Step(`^a song "([^"]*)" with a part "([^"]*)"$`, w.songWithPart)
			sc.Step(`^a song "([^"]*)" with (\d+) parts$`, w.songWithNParts)
			sc.Step(`^I set part "([^"]*)" channel (\d+) to sound bank (\d+) program (\d+) volume (\d+) pan (\d+)$`, w.setSound)
			sc.Step(`^I set part "([^"]*)" channel (\d+) chorus on rate ([0-9.]+) depth (\d+)$`, w.setChorus)
			sc.Step(`^I set part "([^"]*)" channel (\d+) delay on time (\d+) feedback (\d+) mix (\d+)$`, w.setDelay)
			sc.Step(`^I reload the part from the database$`, w.reloadPart)
			sc.Step(`^channel (\d+) sound is bank (\d+) program (\d+) with volume (\d+)$`, w.assertSound)
			sc.Step(`^channel (\d+) chorus is on with rate ([0-9.]+) and depth (\d+)$`, w.assertChorus)
			sc.Step(`^channel (\d+) delay is on with time (\d+) feedback (\d+) mix (\d+)$`, w.assertDelay)
			sc.Step(`^I delete the song "([^"]*)"$`, w.deleteSong)
			sc.Step(`^no parts remain for "([^"]*)"$`, w.noParts)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "persistence.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("persistence feature has failing scenarios")
	}
}
