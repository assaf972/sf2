package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gigsynth/internal/app"
)

// kbPanel is one keyboard's Live strip: selected sound, volume, pan and key
// tuning. References are kept so the panel can refresh from controller state
// (and so tests can assert what is shown).
type kbPanel struct {
	idx        int
	soundLabel *widget.Label
	volLabel   *widget.Label
	panLabel   *widget.Label
	tuneLabel  *widget.Label
}

// buildKeyboardPanels builds the three keyboard strips for the Live view.
func (u *UI) buildKeyboardPanels() fyne.CanvasObject {
	u.kbPanels = make([]*kbPanel, app.MaxKeyboards)
	cards := make([]fyne.CanvasObject, app.MaxKeyboards)
	for i := 0; i < app.MaxKeyboards; i++ {
		p := &kbPanel{
			idx:        i,
			soundLabel: widget.NewLabel("—"),
			volLabel:   widget.NewLabel("0"),
			panLabel:   widget.NewLabel("C"),
			tuneLabel:  widget.NewLabel("+0 st"),
		}
		u.kbPanels[i] = p

		idx := i
		volSlider := widget.NewSlider(0, 127)
		volSlider.OnChanged = func(v float64) {
			if u.suppress {
				return
			}
			u.ctrl.SetLayerVolume(idx, int(v))
			p.volLabel.SetText(fmt.Sprintf("%d", int(v)))
		}
		down := widget.NewButton("−", func() { u.nudgeTranspose(idx, -1) })
		up := widget.NewButton("+", func() { u.nudgeTranspose(idx, +1) })

		body := container.NewVBox(
			widget.NewLabelWithStyle("SOUND", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			p.soundLabel,
			container.NewBorder(nil, nil, widget.NewLabel("Vol"), p.volLabel, volSlider),
			container.NewBorder(nil, nil, widget.NewLabel("Pan"), p.panLabel, widget.NewSlider(0, 127)),
			container.NewBorder(nil, nil, widget.NewLabel("Tune"), container.NewHBox(down, p.tuneLabel, up), nil),
		)
		cards[i] = widget.NewCard(fmt.Sprintf("Keyboard %d", i+1), "", body)
	}
	u.refreshKeyboards()
	return container.NewGridWithColumns(app.MaxKeyboards, cards...)
}

func (u *UI) nudgeTranspose(idx, delta int) {
	sc := u.ctrl.Scene()
	if idx < 0 || idx >= len(sc.Layers) {
		return
	}
	u.ctrl.SetLayerTranspose(idx, sc.Layers[idx].Transpose+delta)
	u.refreshKeyboards()
}

// refreshKeyboards syncs the three panels from the current scene.
func (u *UI) refreshKeyboards() {
	if u.kbPanels == nil {
		return
	}
	sc := u.ctrl.Scene()
	for i, p := range u.kbPanels {
		if i >= len(sc.Layers) {
			continue
		}
		l := sc.Layers[i]
		name := l.PresetName
		if name == "" {
			name = "—"
		}
		p.soundLabel.SetText(name)
		p.volLabel.SetText(fmt.Sprintf("%d", l.Volume))
		p.panLabel.SetText(panText(l.Pan))
		p.tuneLabel.SetText(fmt.Sprintf("%+d st", l.Transpose))
	}
}

func panText(pan int) string {
	switch {
	case pan == 64:
		return "C"
	case pan < 64:
		return fmt.Sprintf("L%d", 64-pan)
	default:
		return fmt.Sprintf("R%d", pan-64)
	}
}
