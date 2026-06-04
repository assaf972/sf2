package ui

import (
	"fmt"
	"path/filepath"
	"testing"

	"gigsynth/internal/app"

	"github.com/cucumber/godog"
)

func TestPiZeroKioskFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var u *UI
			var lib *app.Library
			var vm *pizeroVM

			sc.Step(`^the app runs in pizero mode$`, func() error {
				u, lib = newTestUI(t)
				vm = newPiZeroVM(u.ctrl)
				return nil
			})
			sc.Step(`^a song with (\d+) parts, part (\d+) active$`, func(n, active int) error {
				song, err := lib.DB().CreateSong("Set")
				if err != nil {
					return err
				}
				for i := 1; i <= n; i++ {
					if _, err := lib.DB().AddPart(song.ID, fmt.Sprintf("P%d", i)); err != nil {
						return err
					}
				}
				if err := u.ctrl.LoadSong(song.ID, "Set"); err != nil {
					return err
				}
				u.ctrl.GoToPart(active - 1)
				return nil
			})
			sc.Step(`^the controller sends Program Up$`, func() error { vm.ProgramUp(); return nil })
			sc.Step(`^the controller sends Program Down$`, func() error { vm.ProgramDown(); return nil })
			sc.Step(`^the active part is (\d+)$`, func(want int) error {
				if got := u.ctrl.CurrentPartIndex() + 1; got != want {
					return fmt.Errorf("active part = %d, want %d", got, want)
				}
				return nil
			})
			sc.Step(`^the FX page shows the delay group highlighted$`, func() error {
				vm.fxGroup = "delay"
				return nil
			})
			sc.Step(`^encoder 1 changes to (\d+)$`, func(v int) error { vm.Encoder1(v); return nil })
			sc.Step(`^keyboard 1 delay time is (\d+)$`, func(want int) error {
				if got := u.ctrl.Scene().Layers[0].DelayTime; got != want {
					return fmt.Errorf("delay time = %d, want %d", got, want)
				}
				return nil
			})
			sc.Step(`^I press FX$`, func() error { vm.ToggleFX(); return nil })
			sc.Step(`^the chorus group is highlighted$`, func() error {
				if vm.FXGroup() != "chorus" {
					return fmt.Errorf("fx group = %q, want chorus", vm.FXGroup())
				}
				return nil
			})
			sc.Step(`^the volume knob moves to (\d+)$`, func(v int) error { vm.VolumeKnob(v); return nil })
			sc.Step(`^keyboard 1 volume is (\d+)$`, func(want int) error {
				if got := u.ctrl.Scene().Layers[0].Volume; got != want {
					return fmt.Errorf("volume = %d, want %d", got, want)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{filepath.Join("..", "..", "features", "pizero_kiosk.feature")},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("pizero kiosk feature has failing scenarios")
	}
}
