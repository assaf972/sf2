package ui

import (
	"testing"

	"gigsynth/internal/app"

	"github.com/stretchr/testify/assert"
)

// S04-T02: the Live keyboard panels reflect the controller's scene.
func TestLivePanelsReflectScene(t *testing.T) {
	u, _ := newTestUI(t)
	u.ctrl.SetScene(app.Scene{
		Name:       "t",
		MasterGain: 0.6,
		Layers: []app.Layer{
			{Channel: 0, Enabled: true, PresetName: "Fender Rhodes", Volume: 108, Pan: 64, Transpose: 12},
			{Channel: 1, Enabled: true, PresetName: "Solina", Volume: 96, Pan: 50, Transpose: 0},
			{Channel: 2, Enabled: false, PresetName: "", Volume: 0, Pan: 64, Transpose: 0},
		},
	})

	u.buildKeyboardPanels()
	u.refreshKeyboards()

	assert.Equal(t, "Fender Rhodes", u.kbPanels[0].soundLabel.Text)
	assert.Equal(t, "108", u.kbPanels[0].volLabel.Text)
	assert.Equal(t, "+12 st", u.kbPanels[0].tuneLabel.Text)

	assert.Equal(t, "Solina", u.kbPanels[1].soundLabel.Text)
	assert.Equal(t, "L14", u.kbPanels[1].panLabel.Text) // pan 50 => L14
	assert.Equal(t, "—", u.kbPanels[2].soundLabel.Text) // empty sound shows dash
}
