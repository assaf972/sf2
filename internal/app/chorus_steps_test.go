package app

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

func TestChorusFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var w *fxWorld
			dir := t.TempDir()

			sc.Step(`^keyboard 1 has a chorus insert$`, func() error {
				var err error
				w, err = newFXWorld(dir)
				return err
			})
			sc.Step(`^I enable chorus on keyboard (\d+)$`, func(kb int) error {
				l := w.layer(kb)
				w.c.SetChorus(kb-1, true, l.ChorusRate, l.ChorusDepth)
				return nil
			})
			sc.Step(`^I set chorus rate to ([0-9.]+) and depth to (\d+)$`, func(rate float64, depth int) error {
				l := w.layer(1)
				w.c.SetChorus(0, l.ChorusOn, rate, depth)
				return nil
			})
			sc.Step(`^chorus on keyboard 1 is enabled with rate ([0-9.]+) depth (\d+)$`, func(rate float64, depth int) error {
				w.c.SetChorus(0, true, rate, depth)
				return nil
			})
			sc.Step(`^I save the current part and recall it$`, func() error { return w.saveAndRecall() })
			sc.Step(`^keyboard (\d+) chorus is enabled$`, func(kb int) error {
				if !w.layer(kb).ChorusOn {
					return fmt.Errorf("keyboard %d chorus expected enabled", kb)
				}
				return nil
			})
			sc.Step(`^keyboard (\d+) chorus rate is ([0-9.]+) and depth is (\d+)$`, func(kb int, rate float64, depth int) error {
				return assertChorus(w.layer(kb), true, rate, depth)
			})
			sc.Step(`^keyboard (\d+) chorus is enabled with rate ([0-9.]+) and depth (\d+)$`, func(kb int, rate float64, depth int) error {
				return assertChorus(w.layer(kb), true, rate, depth)
			})
			sc.Step(`^keyboard (\d+) chorus remains disabled$`, func(kb int) error {
				if w.layer(kb).ChorusOn {
					return fmt.Errorf("keyboard %d chorus expected disabled", kb)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "chorus.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("chorus feature has failing scenarios")
	}
}

func assertChorus(l Layer, on bool, rate float64, depth int) error {
	if l.ChorusOn != on || l.ChorusRate != rate || l.ChorusDepth != depth {
		return fmt.Errorf("chorus = {on:%v rate:%v depth:%d}, want {on:%v rate:%v depth:%d}",
			l.ChorusOn, l.ChorusRate, l.ChorusDepth, on, rate, depth)
	}
	return nil
}
