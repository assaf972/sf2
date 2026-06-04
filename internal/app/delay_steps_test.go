package app

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"gigsynth/internal/midiio"

	"github.com/cucumber/godog"
)

func TestDelayFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var w *fxWorld
			dir := t.TempDir()
			const encoder2CC = 71 // M-Audio Enc 2 -> Delay Mix

			sc.Step(`^keyboard 1 has a delay insert$`, func() error {
				var err error
				w, err = newFXWorld(dir)
				return err
			})
			sc.Step(`^I enable delay on keyboard (\d+)$`, func(kb int) error {
				l := w.layer(kb)
				w.c.SetDelay(kb-1, true, l.DelayTime, l.DelayFeedback, l.DelayMix)
				return nil
			})
			sc.Step(`^I set delay time (\d+) ms, feedback (\d+), mix (\d+)$`, func(time, fb, mix int) error {
				l := w.layer(1)
				w.c.SetDelay(0, l.DelayOn, time, fb, mix)
				return nil
			})
			sc.Step(`^delay on keyboard 1 is enabled with time (\d+) feedback (\d+) mix (\d+)$`, func(time, fb, mix int) error {
				w.c.SetDelay(0, true, time, fb, mix)
				return nil
			})
			sc.Step(`^I save the current part and recall it$`, func() error { return w.saveAndRecall() })
			sc.Step(`^keyboard (\d+) delay is enabled$`, func(kb int) error {
				if !w.layer(kb).DelayOn {
					return fmt.Errorf("keyboard %d delay expected enabled", kb)
				}
				return nil
			})
			sc.Step(`^keyboard (\d+) delay time is (\d+), feedback is (\d+), mix is (\d+)$`, func(kb, time, fb, mix int) error {
				return assertDelay(w.layer(kb), time, fb, mix)
			})
			sc.Step(`^keyboard (\d+) delay is enabled with time (\d+) feedback (\d+) mix (\d+)$`, func(kb, time, fb, mix int) error {
				if !w.layer(kb).DelayOn {
					return fmt.Errorf("keyboard %d delay expected enabled", kb)
				}
				return assertDelay(w.layer(kb), time, fb, mix)
			})
			sc.Step(`^keyboard 1 maps encoder 2 to delay mix$`, func() error {
				// Bind keyboard 1 to a device and apply M-Audio (Enc2 -> Delay Mix).
				w.c.SetScene(Scene{Name: "t", MasterGain: 0.6, Layers: []Layer{
					{Channel: 0, Enabled: true, Volume: 100, Pan: 64, Source: "kb1", KeyLow: 0, KeyHigh: 127, DelayOn: true, DelayTime: 300, DelayFeedback: 30, DelayMix: 25},
				}})
				w.c.SetKeyboardMapping(0, "M-Audio")
				return nil
			})
			sc.Step(`^keyboard 1 sends CC for encoder 2 value (\d+)$`, func(val int) error {
				w.c.handleMIDI(midiio.Event{Device: "kb1", Type: midiio.ControlChange, Control: encoder2CC, Value: val})
				return nil
			})
			sc.Step(`^keyboard 1 delay mix is approximately (\d+)$`, func(want int) error {
				got := w.layer(1).DelayMix
				if math.Abs(float64(got-want)) > 1 {
					return fmt.Errorf("delay mix = %d, want ~%d", got, want)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "delay.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("delay feature has failing scenarios")
	}
}

func assertDelay(l Layer, time, fb, mix int) error {
	if !l.DelayOn || l.DelayTime != time || l.DelayFeedback != fb || l.DelayMix != mix {
		return fmt.Errorf("delay = {on:%v time:%d fb:%d mix:%d}, want {on:true time:%d fb:%d mix:%d}",
			l.DelayOn, l.DelayTime, l.DelayFeedback, l.DelayMix, time, fb, mix)
	}
	return nil
}
