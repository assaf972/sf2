package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gigsynth/internal/app"
)

// pizeroVM is the Pi Zero kiosk logic: single keyboard (layer 0), 61-key
// Program/Bank buttons for part/song navigation, four always-live knobs, and an
// FX page whose four encoders bank between the chorus and delay groups.
type pizeroVM struct {
	ctrl    *app.Controller
	fxGroup string // "chorus" | "delay"
}

func newPiZeroVM(c *app.Controller) *pizeroVM { return &pizeroVM{ctrl: c, fxGroup: "delay"} }

// 61-key controller transport.
func (v *pizeroVM) ProgramUp()   { v.ctrl.NextPart() }
func (v *pizeroVM) ProgramDown() { v.ctrl.PrevPart() }

// fxCycle is the order the FX button banks the four encoders through. Starting
// from "delay" the first press lands on "chorus" (kept for the kiosk feature).
var fxCycle = []string{"delay", "chorus", "phaser", "flanger"}

// FX page banking.
func (v *pizeroVM) FXGroup() string { return v.fxGroup }
func (v *pizeroVM) ToggleFX() {
	for i, g := range fxCycle {
		if g == v.fxGroup {
			v.fxGroup = fxCycle[(i+1)%len(fxCycle)]
			return
		}
	}
	v.fxGroup = "chorus"
}

// Encoder1 drives the first parameter of the highlighted FX group.
func (v *pizeroVM) Encoder1(value int) {
	l := v.layer0()
	switch v.fxGroup {
	case "delay":
		v.ctrl.SetDelay(0, true, value, l.DelayFeedback, l.DelayMix) // value = time ms
	case "phaser":
		v.ctrl.SetPhaser(0, true, float64(value), l.PhaserDepth, l.PhaserFeedback) // value = rate
	case "flanger":
		v.ctrl.SetFlanger(0, true, float64(value), l.FlangerDepth, l.FlangerFeedback, l.FlangerMix) // value = rate
	default: // chorus
		v.ctrl.SetChorus(0, true, float64(value), l.ChorusDepth) // value = rate
	}
}

// VolumeKnob is always live regardless of which screen is shown.
func (v *pizeroVM) VolumeKnob(value int) { v.ctrl.SetLayerVolume(0, value) }

func (v *pizeroVM) layer0() app.Layer {
	sc := v.ctrl.Scene()
	if len(sc.Layers) == 0 {
		return app.Layer{}
	}
	return sc.Layers[0]
}

// buildPiZero renders the compact kiosk root content.
func (u *UI) buildPiZero() fyne.CanvasObject {
	vm := newPiZeroVM(u.ctrl)
	status := widget.NewLabel("")
	refresh := func() {
		status.SetText(fmt.Sprintf("%s · %s (%d/%d)",
			u.ctrl.CurrentSong(), u.ctrl.CurrentPartName(),
			u.ctrl.CurrentPartIndex()+1, max1(u.ctrl.PartCount())))
	}
	refresh()
	prev := widget.NewButton("◀ Prev Part", func() { vm.ProgramDown(); refresh() })
	next := widget.NewButton("Next Part ▶", func() { vm.ProgramUp(); refresh() })
	fxLabel := widget.NewLabel("FX group: " + vm.FXGroup())
	fxToggle := widget.NewButton("FX ▸ next (chorus·phaser·flanger·delay)", func() {
		vm.ToggleFX()
		fxLabel.SetText("FX group: " + vm.FXGroup())
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("GIGSYNTH", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		status,
		container.NewGridWithColumns(2, prev, next),
		fxLabel,
		fxToggle,
	)
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
