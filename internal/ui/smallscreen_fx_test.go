package ui

import (
	"testing"

	"gigsynth/internal/app"

	"github.com/stretchr/testify/assert"
)

// S25-T01: the 7-inch layout surfaces the Phaser and Flanger values for the
// focused keyboard alongside chorus/delay.
func TestTouch7ShowsPhaserFlanger(t *testing.T) {
	u, _ := newTestUI(t)
	u.ctrl.SetScene(app.Scene{
		Name: "t", MasterGain: 0.6,
		Layers: []app.Layer{
			{Channel: 0, Enabled: true, PresetName: "Pad",
				PhaserRate: 0.5, PhaserDepth: 60, PhaserFeedback: 30,
				FlangerRate: 0.25, FlangerDepth: 70, FlangerFeedback: 40, FlangerMix: 55},
		},
	})
	u.buildTouch7()
	u.t7.focus(0)
	assert.Equal(t, "0.5 Hz", u.t7.phRate.Text)
	assert.Equal(t, "60", u.t7.phDepth.Text)
	assert.Equal(t, "55", u.t7.flMix.Text)
}

// S25-T02: the Pi Zero FX button banks the four encoders through all four
// effects (delay → chorus → phaser → flanger), and encoder 1 drives the
// highlighted effect.
func TestPiZeroBanksThroughAllFourEffects(t *testing.T) {
	u, _ := newTestUI(t)
	u.ctrl.SetScene(app.Scene{Name: "t", MasterGain: 0.6,
		Layers: []app.Layer{{Channel: 0, Enabled: true}}})
	vm := newPiZeroVM(u.ctrl)

	assert.Equal(t, "delay", vm.FXGroup())
	vm.ToggleFX()
	assert.Equal(t, "chorus", vm.FXGroup(), "first press lands on chorus (kiosk feature expects this)")
	vm.ToggleFX()
	assert.Equal(t, "phaser", vm.FXGroup())
	vm.Encoder1(3) // phaser rate
	assert.True(t, u.ctrl.Scene().Layers[0].PhaserOn)
	assert.Equal(t, 3.0, u.ctrl.Scene().Layers[0].PhaserRate)

	vm.ToggleFX()
	assert.Equal(t, "flanger", vm.FXGroup())
	vm.Encoder1(2) // flanger rate
	assert.True(t, u.ctrl.Scene().Layers[0].FlangerOn)
	assert.Equal(t, 2.0, u.ctrl.Scene().Layers[0].FlangerRate)

	vm.ToggleFX()
	assert.Equal(t, "delay", vm.FXGroup(), "cycle wraps back to delay")
}
