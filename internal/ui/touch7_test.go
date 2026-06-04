package ui

import (
	"testing"

	"gigsynth/internal/app"

	"github.com/stretchr/testify/assert"
)

// S15-T01: the 7-inch layout focuses one keyboard and shows that keyboard's
// sound and FX values.
func TestTouch7FocusesKeyboard(t *testing.T) {
	u, _ := newTestUI(t)
	u.ctrl.SetScene(app.Scene{
		Name:       "t",
		MasterGain: 0.6,
		Layers: []app.Layer{
			{Channel: 0, Enabled: true, PresetName: "Fender Rhodes", ChorusDepth: 40, DelayMix: 20},
			{Channel: 1, Enabled: true, PresetName: "Mellotron Choir", ChorusRate: 1.2, ChorusDepth: 70, DelayTime: 480, DelayFeedback: 50, DelayMix: 45},
		},
	})
	u.buildTouch7()

	u.t7.focus(1) // focus Keyboard 2
	assert.Equal(t, "Keyboard 2", u.t7.title.Text)
	assert.Equal(t, "Mellotron Choir", u.t7.sound.Text)
	assert.Equal(t, "70", u.t7.chDepth.Text)
	assert.Equal(t, "480 ms", u.t7.dlTime.Text)
	assert.Equal(t, "45", u.t7.dlMix.Text)

	u.t7.focus(0) // back to Keyboard 1
	assert.Equal(t, "Fender Rhodes", u.t7.sound.Text)
	assert.Equal(t, "40", u.t7.chDepth.Text)
}
