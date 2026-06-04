package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gigsynth/internal/app"
)

// touch7 is the 7-inch (800x480) layout: a focus-one-keyboard view so the five
// chorus/delay knobs stay finger-sized. KB1/KB2/KB3 tabs switch the focused
// keyboard; its sound, mix and full FX are shown large.
type touch7 struct {
	u       *UI
	focused int
	title   *widget.Label
	sound   *widget.Label
	chRate  *widget.Label
	chDepth *widget.Label
	dlTime  *widget.Label
	dlFb    *widget.Label
	dlMix   *widget.Label
	phRate  *widget.Label
	phDepth *widget.Label
	phFb    *widget.Label
	flRate  *widget.Label
	flDepth *widget.Label
	flFb    *widget.Label
	flMix   *widget.Label
}

func (u *UI) buildTouch7() fyne.CanvasObject {
	t := &touch7{
		u:       u,
		title:   widget.NewLabel("Keyboard 1"),
		sound:   widget.NewLabel("—"),
		chRate:  widget.NewLabel("-"),
		chDepth: widget.NewLabel("-"),
		dlTime:  widget.NewLabel("-"),
		dlFb:    widget.NewLabel("-"),
		dlMix:   widget.NewLabel("-"),
		phRate:  widget.NewLabel("-"),
		phDepth: widget.NewLabel("-"),
		phFb:    widget.NewLabel("-"),
		flRate:  widget.NewLabel("-"),
		flDepth: widget.NewLabel("-"),
		flFb:    widget.NewLabel("-"),
		flMix:   widget.NewLabel("-"),
	}
	u.t7 = t

	tabs := make([]fyne.CanvasObject, app.MaxKeyboards)
	for i := 0; i < app.MaxKeyboards; i++ {
		i := i
		tabs[i] = widget.NewButton(fmt.Sprintf("KB%d", i+1), func() { t.focus(i) })
	}

	chorusDelay := container.NewGridWithColumns(5,
		labeled("Ch Rate", t.chRate), labeled("Ch Depth", t.chDepth),
		labeled("Dl Time", t.dlTime), labeled("Dl F.Back", t.dlFb), labeled("Dl Mix", t.dlMix),
	)
	phaserFlanger := container.NewGridWithColumns(5,
		labeled("Ph Rate", t.phRate), labeled("Ph Depth", t.phDepth), labeled("Ph F.Back", t.phFb),
		labeled("Fl Rate", t.flRate), labeled("Fl Depth", t.flDepth),
	)
	phaserFlanger2 := container.NewGridWithColumns(5,
		labeled("Fl F.Back", t.flFb), labeled("Fl Mix", t.flMix),
	)
	t.focus(0)
	return container.NewVBox(
		container.NewGridWithColumns(app.MaxKeyboards, tabs...),
		widget.NewCard("", "", container.NewVBox(t.title, t.sound)),
		widget.NewCard("Chorus + Delay", "", chorusDelay),
		widget.NewCard("Phaser + Flanger", "", container.NewVBox(phaserFlanger, phaserFlanger2)),
	)
}

func labeled(name string, v *widget.Label) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(name), v)
}

func (t *touch7) focus(i int) {
	t.focused = i
	t.refresh()
}

func (t *touch7) refresh() {
	sc := t.u.ctrl.Scene()
	if t.focused < 0 || t.focused >= len(sc.Layers) {
		return
	}
	l := sc.Layers[t.focused]
	t.title.SetText(fmt.Sprintf("Keyboard %d", t.focused+1))
	name := l.PresetName
	if name == "" {
		name = "—"
	}
	t.sound.SetText(name)
	t.chRate.SetText(fmt.Sprintf("%.1f Hz", l.ChorusRate))
	t.chDepth.SetText(fmt.Sprintf("%d", l.ChorusDepth))
	t.dlTime.SetText(fmt.Sprintf("%d ms", l.DelayTime))
	t.dlFb.SetText(fmt.Sprintf("%d", l.DelayFeedback))
	t.dlMix.SetText(fmt.Sprintf("%d", l.DelayMix))
	t.phRate.SetText(fmt.Sprintf("%.1f Hz", l.PhaserRate))
	t.phDepth.SetText(fmt.Sprintf("%d", l.PhaserDepth))
	t.phFb.SetText(fmt.Sprintf("%d", l.PhaserFeedback))
	t.flRate.SetText(fmt.Sprintf("%.1f Hz", l.FlangerRate))
	t.flDepth.SetText(fmt.Sprintf("%d", l.FlangerDepth))
	t.flFb.SetText(fmt.Sprintf("%d", l.FlangerFeedback))
	t.flMix.SetText(fmt.Sprintf("%d", l.FlangerMix))
}
