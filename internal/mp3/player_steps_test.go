package mp3

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

func TestMP3PlayerFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var p *Player

			sc.Step(`^a loaded MP3 file of known duration$`, func() error {
				p = New(nil) // headless: no audio backend
				p.Load(332.0)
				return nil
			})
			sc.Step(`^I press Start$`, func() error { p.Start(); return nil })
			sc.Step(`^I press Stop$`, func() error { p.Stop(); return nil })
			sc.Step(`^looping is enabled$`, func() error { p.SetLoop(true); return nil })
			sc.Step(`^playback reaches the end$`, func() error { p.ReachEnd(); return nil })
			sc.Step(`^I set pitch to (\d+) semitones$`, func(st int) error { p.SetPitch(st); return nil })
			sc.Step(`^the player state is "([^"]*)"$`, func(want string) error {
				if got := p.State().String(); got != want {
					return fmt.Errorf("state = %q, want %q", got, want)
				}
				return nil
			})
			sc.Step(`^the playhead is at 0$`, func() error {
				if p.Position() != 0 {
					return fmt.Errorf("playhead = %v, want 0", p.Position())
				}
				return nil
			})
			sc.Step(`^playback restarts from 0$`, func() error {
				if p.Position() != 0 {
					return fmt.Errorf("playhead = %v, want 0 after loop", p.Position())
				}
				return nil
			})
			sc.Step(`^the resample ratio is approximately ([0-9.]+)$`, func(want float64) error {
				if math.Abs(p.Ratio()-want) > 0.001 {
					return fmt.Errorf("ratio = %v, want ~%v", p.Ratio(), want)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "mp3_player.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("mp3 player feature has failing scenarios")
	}
}
