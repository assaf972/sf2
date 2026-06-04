// Package ui builds the Fyne user interface: a per-layer mixer, MIDI device
// management, scene save/recall, and an on-screen test keyboard.
package ui

import (
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"gigsynth/internal/app"
	"gigsynth/internal/db"
	"gigsynth/internal/engine"
	"gigsynth/internal/midiio"
	"gigsynth/internal/recording"
)

type UI struct {
	ctrl  *app.Controller
	midi  *midiio.Manager
	store *app.Store
	lib   *app.Library
	win   fyne.Window
	fapp  fyne.App

	suppress bool

	strips        []*strip
	sfLabel       *widget.Label
	sceneSelect   *widget.Select
	masterSlider  *widget.Slider
	deviceBox     *fyne.Container
	presetOptions []string
	presetByLabel map[string]engine.Preset

	noteOptions      []string
	transposeOptions []string

	// shell / songs / settings widgets
	partLabel        *widget.Label
	songList         *widget.List
	songs            []db.Song
	partList         *widget.Label
	selectedSong     int64
	selectedSongName string
	settingsSFLabel  *widget.Label
	kbPanels         []*kbPanel
	t7               *touch7

	// recordings view + Live REC toggle
	recList    *widget.List
	recStatus  *widget.Label
	recToggle  *widget.Button
	recordings []recording.Recording
	recSel     int
}

type strip struct {
	idx       int
	name      *widget.Entry
	enabled   *widget.Check
	preset    *widget.Select
	volume    *widget.Slider
	volLabel  *widget.Label
	pan       *widget.Slider
	mute      *widget.Button
	source    *widget.Select
	low       *widget.Select
	high      *widget.Select
	transpose *widget.Select
}

// Run builds and shows the window, blocking until it closes. mode selects the
// layout: "" (desktop 3-view shell), "touch7" (7-inch focus layout), or
// "pizero" (small-LCD kiosk).
func Run(ctrl *app.Controller, m *midiio.Manager, store *app.Store, lib *app.Library, mode string) {
	a := fyneapp.NewWithID("com.gigsynth.app")
	w := a.NewWindow("GigSynth — FluidSynth Live Rig")

	u := &UI{
		ctrl:          ctrl,
		midi:          m,
		store:         store,
		lib:           lib,
		win:           w,
		fapp:          a,
		presetByLabel: map[string]engine.Preset{},
	}
	// Note-number options: index == MIDI note number.
	for n := 0; n <= 127; n++ {
		u.noteOptions = append(u.noteOptions, fmt.Sprintf("%s (%d)", app.NoteName(n), n))
	}
	for t := -24; t <= 24; t++ {
		u.transposeOptions = append(u.transposeOptions, fmt.Sprintf("%+d", t))
	}

	switch mode {
	case "pizero":
		w.SetContent(u.buildPiZero())
		w.Resize(fyne.NewSize(480, 320))
	case "touch7":
		w.SetContent(u.buildTouch7())
		w.Resize(fyne.NewSize(800, 480))
	default:
		ctrl.SetOnChange(func() { fyne.Do(u.refresh) })
		w.SetContent(u.build())
		w.Resize(fyne.NewSize(1180, 760))
		u.refresh()
	}
	w.ShowAndRun()
}

func (u *UI) build() fyne.CanvasObject {
	return u.buildShell()
}

// liveContent is the Live tab: top bar (master/scenes/PANIC), the part shuttle,
// the mixer strips, and the device/test-keyboard panel.
func (u *UI) liveContent() fyne.CanvasObject {
	return container.NewBorder(
		container.NewVBox(u.buildTopBar(), u.buildPartShuttle(), u.buildKeyboardPanels()),
		u.buildBottom(),
		nil, nil,
		u.buildMixer(),
	)
}

// ---- top bar: soundfont, master gain, scenes, panic ----

func (u *UI) buildTopBar() fyne.CanvasObject {
	loadBtn := widget.NewButton("Load SoundFont…", u.onLoadSoundFont)
	u.sfLabel = widget.NewLabel("(no SoundFont loaded)")

	u.masterSlider = widget.NewSlider(0, 2)
	u.masterSlider.Step = 0.01
	u.masterSlider.OnChanged = func(v float64) {
		if u.suppress {
			return
		}
		u.ctrl.SetMasterGain(v)
	}
	masterLabel := widget.NewLabel("Master")

	panic := widget.NewButton("PANIC", func() { u.ctrl.Panic() })
	panic.Importance = widget.DangerImportance

	u.sceneSelect = widget.NewSelect(nil, nil) // populated in refresh
	u.sceneSelect.PlaceHolder = "(scenes)"
	recallBtn := widget.NewButton("Recall", u.onRecallScene)
	saveBtn := widget.NewButton("Save As…", u.onSaveScene)

	row1 := container.NewBorder(nil, nil,
		loadBtn, nil,
		u.sfLabel,
	)
	row2 := container.NewBorder(nil, nil,
		container.NewHBox(masterLabel),
		container.NewHBox(widget.NewLabel("Scene:"), u.sceneSelect, recallBtn, saveBtn, u.buildRecordToggle(), panic),
		u.masterSlider,
	)
	return container.NewVBox(row1, widget.NewSeparator(), row2, widget.NewSeparator())
}

// ---- center: mixer strips ----

func (u *UI) buildMixer() fyne.CanvasObject {
	scene := u.ctrl.Scene()
	cols := make([]fyne.CanvasObject, 0, len(scene.Layers))
	u.strips = make([]*strip, len(scene.Layers))
	for i := range scene.Layers {
		s := u.buildStrip(i)
		u.strips[i] = s
		cols = append(cols, u.stripCard(s))
	}
	return container.NewGridWithColumns(len(cols), cols...)
}

func (u *UI) buildStrip(i int) *strip {
	s := &strip{idx: i}

	s.name = widget.NewEntry()
	s.name.OnChanged = func(v string) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerName(i, v)
	}

	s.enabled = widget.NewCheck("On", func(b bool) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerEnabled(i, b)
	})

	s.preset = widget.NewSelect(nil, func(label string) {
		if u.suppress {
			return
		}
		if p, ok := u.presetByLabel[label]; ok {
			u.ctrl.SetLayerPreset(i, p.Bank, p.Program, p.Name)
		}
	})
	s.preset.PlaceHolder = "(load a SoundFont)"

	s.volume = widget.NewSlider(0, 127)
	s.volLabel = widget.NewLabel("100")
	s.volume.OnChanged = func(v float64) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerVolume(i, int(v))
		s.volLabel.SetText(fmt.Sprintf("%d", int(v)))
	}

	s.pan = widget.NewSlider(0, 127)
	s.pan.OnChanged = func(v float64) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerPan(i, int(v))
	}

	s.mute = widget.NewButton("Mute", func() {
		sc := u.ctrl.Scene()
		u.ctrl.SetLayerMute(i, !sc.Layers[i].Mute)
	})

	s.source = widget.NewSelect(nil, func(v string) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerSource(i, v)
	})

	s.low = widget.NewSelect(u.noteOptions, func(string) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerRange(i, s.low.SelectedIndex(), s.high.SelectedIndex())
	})
	s.high = widget.NewSelect(u.noteOptions, func(string) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerRange(i, s.low.SelectedIndex(), s.high.SelectedIndex())
	})

	s.transpose = widget.NewSelect(u.transposeOptions, func(string) {
		if u.suppress {
			return
		}
		u.ctrl.SetLayerTranspose(i, s.transpose.SelectedIndex()-24)
	})

	return s
}

func (u *UI) stripCard(s *strip) fyne.CanvasObject {
	header := container.NewBorder(nil, nil, s.enabled, s.mute, s.name)

	mix := container.NewVBox(
		widget.NewLabel("Sound"),
		s.preset,
		container.NewBorder(nil, nil, widget.NewLabel("Vol"), s.volLabel, s.volume),
		container.NewBorder(nil, nil, widget.NewLabel("Pan"), nil, s.pan),
	)

	routing := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabel("Routing"),
		container.NewBorder(nil, nil, widget.NewLabel("From"), nil, s.source),
		container.NewGridWithColumns(2,
			container.NewBorder(nil, nil, widget.NewLabel("Low"), nil, s.low),
			container.NewBorder(nil, nil, widget.NewLabel("High"), nil, s.high),
		),
		container.NewBorder(nil, nil, widget.NewLabel("Transpose"), nil, s.transpose),
	)

	return widget.NewCard("", "", container.NewVBox(header, widget.NewSeparator(), mix, routing))
}

// ---- bottom: devices + test keyboard ----

func (u *UI) buildBottom() fyne.CanvasObject {
	u.deviceBox = container.NewVBox()
	refresh := widget.NewButton("Rescan MIDI", u.refreshDevices)
	u.refreshDevices()

	devPanel := widget.NewCard("MIDI Keyboards", "",
		container.NewVBox(refresh, u.deviceBox))

	kb := u.buildKeyboard()
	kbPanel := widget.NewCard("Test Keyboard", "click keys to audition the patch", kb)

	return container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(2, devPanel, kbPanel),
	)
}

func (u *UI) refreshDevices() {
	u.deviceBox.RemoveAll()
	if u.midi == nil {
		u.deviceBox.Add(widget.NewLabel("MIDI unavailable"))
		u.deviceBox.Refresh()
		return
	}
	inputs, err := u.midi.ListInputs()
	if err != nil {
		u.deviceBox.Add(widget.NewLabel("MIDI error: " + err.Error()))
		u.deviceBox.Refresh()
		return
	}
	if len(inputs) == 0 {
		u.deviceBox.Add(widget.NewLabel("No MIDI inputs found. Connect a keyboard and Rescan."))
	}
	connected := map[string]bool{}
	for _, c := range u.midi.Connected() {
		connected[c] = true
	}
	for _, name := range inputs {
		name := name
		chk := widget.NewCheck(name, func(b bool) {
			if b {
				if err := u.midi.Connect(name); err != nil {
					dialog.ShowError(err, u.win)
				}
			} else {
				u.midi.Disconnect(name)
			}
			u.refreshSources()
		})
		chk.SetChecked(connected[name])
		u.deviceBox.Add(chk)
	}
	u.deviceBox.Refresh()
	u.refreshSources()
}

// buildKeyboard renders two octaves of clickable keys (C3..B4).
func (u *UI) buildKeyboard() fyne.CanvasObject {
	const low, high = 48, 72 // C3..C5
	keys := make([]fyne.CanvasObject, 0, high-low+1)
	for n := low; n <= high; n++ {
		n := n
		b := widget.NewButton(app.NoteName(n), func() {
			u.ctrl.VirtualNoteOn(n, 100)
			time.AfterFunc(600*time.Millisecond, func() {
				u.ctrl.VirtualNoteOff(n)
			})
		})
		keys = append(keys, b)
	}
	return container.NewGridWithColumns(13, keys...)
}

// ---- actions ----

func (u *UI) onLoadSoundFont() {
	fd := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		path := rc.URI().Path()
		rc.Close()
		if err := u.ctrl.LoadSoundFont(path); err != nil {
			dialog.ShowError(err, u.win)
			return
		}
	}, u.win)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".sf2", ".sf3"}))
	fd.Show()
}

func (u *UI) onSaveScene() {
	entry := widget.NewEntry()
	entry.SetText(u.ctrl.Scene().Name)
	dialog.ShowForm("Save Scene", "Save", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(ok bool) {
			if !ok {
				return
			}
			sc := u.ctrl.Scene()
			sc.Name = entry.Text
			if err := u.store.Save(sc); err != nil {
				dialog.ShowError(err, u.win)
				return
			}
			u.refreshScenes()
			u.sceneSelect.SetSelected(sc.Name)
		}, u.win)
}

func (u *UI) onRecallScene() {
	name := u.sceneSelect.Selected
	if name == "" {
		return
	}
	sc, err := u.store.Load(name)
	if err != nil {
		dialog.ShowError(err, u.win)
		return
	}
	u.ctrl.SetScene(sc)
}

// ---- refresh: sync widgets from controller state ----

func (u *UI) refresh() {
	u.suppress = true
	defer func() { u.suppress = false }()

	scene := u.ctrl.Scene()
	presets := u.ctrl.Presets()

	// SoundFont label.
	if p := u.ctrl.SoundFontPath(); p != "" {
		u.sfLabel.SetText("SoundFont: " + filepath.Base(p))
	} else {
		u.sfLabel.SetText("(no SoundFont loaded)")
	}

	// Preset options.
	u.presetOptions = u.presetOptions[:0]
	u.presetByLabel = map[string]engine.Preset{}
	for _, p := range presets {
		label := fmt.Sprintf("%03d:%03d  %s", p.Bank, p.Program, p.Name)
		u.presetOptions = append(u.presetOptions, label)
		u.presetByLabel[label] = p
	}

	u.masterSlider.SetValue(scene.MasterGain)

	for i, l := range scene.Layers {
		if i >= len(u.strips) {
			break
		}
		s := u.strips[i]
		s.name.SetText(l.Name)
		s.enabled.SetChecked(l.Enabled)
		s.volume.SetValue(float64(l.Volume))
		s.volLabel.SetText(fmt.Sprintf("%d", l.Volume))
		s.pan.SetValue(float64(l.Pan))
		if l.Mute {
			s.mute.SetText("Muted")
			s.mute.Importance = widget.WarningImportance
		} else {
			s.mute.SetText("Mute")
			s.mute.Importance = widget.MediumImportance
		}
		s.mute.Refresh()

		s.preset.Options = append([]string(nil), u.presetOptions...)
		if cur := presetLabel(presets, l.Bank, l.Program); cur != "" {
			s.preset.SetSelected(cur)
		} else {
			s.preset.ClearSelected()
		}
		s.preset.Refresh()

		if l.KeyLow >= 0 && l.KeyLow < len(u.noteOptions) {
			s.low.SetSelectedIndex(l.KeyLow)
		}
		if l.KeyHigh >= 0 && l.KeyHigh < len(u.noteOptions) {
			s.high.SetSelectedIndex(l.KeyHigh)
		}
		ti := l.Transpose + 24
		if ti >= 0 && ti < len(u.transposeOptions) {
			s.transpose.SetSelectedIndex(ti)
		}
	}

	u.refreshSources()
	u.refreshScenes()
	u.refreshKeyboards()
	u.refreshShuttle()
	u.refreshSongs()
	if u.settingsSFLabel != nil {
		u.settingsSFLabel.SetText(u.sfText())
	}
}

func (u *UI) refreshSources() {
	opts := []string{app.SourceAny}
	if u.midi != nil {
		opts = append(opts, u.midi.Connected()...)
	}
	scene := u.ctrl.Scene()
	for i, s := range u.strips {
		s.source.Options = opts
		sel := scene.Layers[i].Source
		found := false
		for _, o := range opts {
			if o == sel {
				found = true
				break
			}
		}
		if !found {
			// Keep the saved name visible even if that device isn't present.
			s.source.Options = append(opts, sel)
		}
		s.source.SetSelected(sel)
		s.source.Refresh()
	}
}

func (u *UI) refreshScenes() {
	names, err := u.store.List()
	if err != nil {
		return
	}
	sel := u.sceneSelect.Selected
	u.sceneSelect.Options = names
	u.sceneSelect.Refresh()
	if sel != "" {
		u.sceneSelect.SetSelected(sel)
	}
}

func presetLabel(presets []engine.Preset, bank, program int) string {
	for _, p := range presets {
		if p.Bank == bank && p.Program == program {
			return fmt.Sprintf("%03d:%03d  %s", p.Bank, p.Program, p.Name)
		}
	}
	return ""
}
