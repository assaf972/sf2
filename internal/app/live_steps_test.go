package app

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

func TestLiveMixerFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var c *Controller
			var f *FakeEngine

			sc.Step(`^a loaded SoundFont$`, func() error {
				f = &FakeEngine{}
				c = NewController(f, nil)
				return c.LoadSoundFont("x.sf2")
			})
			sc.Step(`^keyboard 1 is enabled with sound "([^"]*)"$`, func(sound string) error {
				c.SetScene(Scene{Name: "t", MasterGain: 0.6, Layers: []Layer{
					{Channel: 0, Enabled: true, Volume: 100, Pan: 64, PresetName: sound, Source: SourceAny, KeyLow: 0, KeyHigh: 127},
				}})
				f.Reset()
				return nil
			})
			sc.Step(`^I set keyboard 1 volume to (\d+)$`, func(v int) error { c.SetLayerVolume(0, v); return nil })
			sc.Step(`^I set keyboard 1 pan to centre$`, func() error { c.SetLayerPan(0, 64); return nil })
			sc.Step(`^keyboard 1 key tuning is \+(\d+) semitones$`, func(st int) error { c.SetLayerTranspose(0, st); return nil })
			sc.Step(`^keyboard 1 plays note (\d+)$`, func(n int) error { c.VirtualNoteOn(n, 100); return nil })
			sc.Step(`^the synth receives CC7 value (\d+) on channel (\d+)$`, func(val, ch int) error {
				return mustContain(f, fmt.Sprintf("CC ch=%d ctrl=7 val=%d", ch, val))
			})
			sc.Step(`^the synth receives CC10 value (\d+) on channel (\d+)$`, func(val, ch int) error {
				return mustContain(f, fmt.Sprintf("CC ch=%d ctrl=10 val=%d", ch, val))
			})
			sc.Step(`^the synth receives NoteOn key (\d+) on channel (\d+)$`, func(key, ch int) error {
				return mustContain(f, fmt.Sprintf("NoteOn ch=%d key=%d vel=100", ch, key))
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "live_mixer.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("live mixer feature has failing scenarios")
	}
}

func mustContain(f *FakeEngine, want string) error {
	for _, c := range f.Calls {
		if c == want {
			return nil
		}
	}
	return fmt.Errorf("expected engine call %q in %v", want, f.Calls)
}
