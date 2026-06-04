package app

import (
	"fmt"
	"path/filepath"
	"testing"

	"gigsynth/internal/midiio"

	"github.com/cucumber/godog"
)

func TestMidiMappingFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var c *Controller
			var f *FakeEngine
			device := "kb1"
			var startVol int

			sc.Step(`^keyboard 1 uses the "([^"]*)" mapping preset$`, func(preset string) error {
				f = &FakeEngine{}
				c = NewController(f, nil)
				// Bind layer 0 to device "kb1" and apply the chosen preset to slot 0.
				c.SetScene(Scene{Name: "t", MasterGain: 0.6, Layers: []Layer{
					{Channel: 0, Enabled: true, Volume: 100, Pan: 64, Source: device, KeyLow: 0, KeyHigh: 127, ChorusRate: 0.8, ChorusDepth: 50},
				}})
				c.SetKeyboardMapping(0, preset)
				startVol = c.Scene().Layers[0].Volume
				return nil
			})
			sc.Step(`^keyboard 1 drives layer 0$`, func() error { return nil })
			sc.Step(`^keyboard 1 sends CC (\d+) value (\d+)$`, func(cc, val int) error {
				c.handleMIDI(midiio.Event{Device: device, Type: midiio.ControlChange, Control: cc, Value: val})
				return nil
			})
			sc.Step(`^layer 0 volume becomes (\d+)$`, func(want int) error {
				if got := c.Scene().Layers[0].Volume; got != want {
					return fmt.Errorf("volume = %d, want %d", got, want)
				}
				return nil
			})
			sc.Step(`^layer 0 chorus depth becomes (\d+)$`, func(want int) error {
				if got := c.Scene().Layers[0].ChorusDepth; got != want {
					return fmt.Errorf("chorus depth = %d, want %d", got, want)
				}
				return nil
			})
			sc.Step(`^layer 0 volume is unchanged$`, func() error {
				if got := c.Scene().Layers[0].Volume; got != startVol {
					return fmt.Errorf("volume changed to %d (was %d)", got, startVol)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "midi_mapping.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("midi mapping feature has failing scenarios")
	}
}
