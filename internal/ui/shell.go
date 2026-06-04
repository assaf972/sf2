package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"gigsynth/internal/app"
	"gigsynth/internal/db"
	"gigsynth/internal/midimap"
)

// shellTabLive/Songs/Settings are the three top-level view names.
const (
	shellTabLive     = "Live"
	shellTabSongs    = "Songs"
	shellTabSettings = "Settings"
)

// buildShell composes the three-view application shell.
func (u *UI) buildShell() *container.AppTabs {
	tabs := container.NewAppTabs(
		container.NewTabItem(shellTabLive, u.liveContent()),
		container.NewTabItem(shellTabSongs, u.buildSongsView()),
		container.NewTabItem(shellTabSettings, u.buildSettingsView()),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	return tabs
}

// ---- Live: part shuttle ----

func (u *UI) buildPartShuttle() fyne.CanvasObject {
	u.partLabel = widget.NewLabel("(no song loaded)")
	prev := widget.NewButton("◀ Part", func() {
		u.ctrl.PrevPart()
		u.refreshShuttle()
	})
	next := widget.NewButton("Part ▶", func() {
		u.ctrl.NextPart()
		u.refreshShuttle()
	})
	u.refreshShuttle()
	return container.NewBorder(nil, nil, prev, next, container.NewCenter(u.partLabel))
}

func (u *UI) refreshShuttle() {
	if u.partLabel == nil {
		return
	}
	n := u.ctrl.PartCount()
	if n == 0 {
		u.partLabel.SetText("(no song loaded)")
		return
	}
	u.partLabel.SetText(fmt.Sprintf("%s · %s  (%d/%d)",
		u.ctrl.CurrentSong(), u.ctrl.CurrentPartName(), u.ctrl.CurrentPartIndex()+1, n))
}

// ---- Songs view ----

func (u *UI) buildSongsView() fyne.CanvasObject {
	u.songList = widget.NewList(
		func() int { return len(u.songs) },
		func() fyne.CanvasObject { return widget.NewLabel("song") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(u.songs[i].Name)
		},
	)
	u.songList.OnSelected = func(i widget.ListItemID) {
		if i >= 0 && i < len(u.songs) {
			u.loadSongParts(u.songs[i])
		}
	}
	u.partList = widget.NewLabel("Select a song to see its parts.")

	newBtn := widget.NewButton("+ New Song", func() {
		entry := widget.NewEntry()
		dialog.ShowForm("New Song", "Create", "Cancel",
			[]*widget.FormItem{widget.NewFormItem("Name", entry)},
			func(ok bool) {
				if !ok || u.lib == nil {
					return
				}
				if _, err := u.lib.DB().CreateSong(entry.Text); err != nil {
					dialog.ShowError(err, u.win)
					return
				}
				u.refreshSongs()
			}, u.win)
	})
	recallBtn := widget.NewButton("Load Song", func() {
		if u.selectedSong > 0 {
			_ = u.ctrl.LoadSong(u.selectedSong, u.selectedSongName)
			u.refreshShuttle()
		}
	})

	u.refreshSongs()
	left := container.NewBorder(nil, container.NewVBox(newBtn, recallBtn), nil, nil, u.songList)
	return container.NewHSplit(left, container.NewScroll(u.partList))
}

func (u *UI) refreshSongs() {
	if u.lib == nil || u.songList == nil {
		return
	}
	songs, err := u.lib.DB().ListSongs()
	if err != nil {
		return
	}
	u.songs = songs
	u.songList.Refresh()
}

func (u *UI) loadSongParts(s db.Song) {
	u.selectedSong = s.ID
	u.selectedSongName = s.Name
	if u.lib == nil {
		return
	}
	parts, err := u.lib.DB().ListParts(s.ID)
	if err != nil {
		return
	}
	txt := fmt.Sprintf("Parts of %q:\n", s.Name)
	for _, p := range parts {
		txt += fmt.Sprintf("  %d. %s\n", p.SortOrder, p.Name)
	}
	u.partList.SetText(txt)
}

// ---- Settings view ----

func (u *UI) buildSettingsView() fyne.CanvasObject {
	u.settingsSFLabel = widget.NewLabel(u.sfText())
	loadBtn := widget.NewButton("Load SoundFont…", u.onLoadSoundFont)

	mapping := container.NewVBox()
	for kb := 0; kb < app.MaxKeyboards; kb++ {
		kb := kb
		sel := widget.NewSelect(midimap.Names(), func(name string) {
			u.ctrl.SetKeyboardMapping(kb, name)
		})
		sel.SetSelected("Generic GM")
		mapping.Add(container.NewBorder(nil, nil,
			widget.NewLabel(fmt.Sprintf("Keyboard %d mapping", kb+1)), nil, sel))
	}

	audio := widget.NewLabel(u.audioText())

	return container.NewVBox(
		widget.NewCard("SoundFont", "", container.NewVBox(u.settingsSFLabel, loadBtn)),
		widget.NewCard("MIDI Keyboards & Mapping", "", mapping),
		widget.NewCard("Audio Engine", "", audio),
	)
}

func (u *UI) sfText() string {
	if p := u.ctrl.SoundFontPath(); p != "" {
		return "SoundFont: " + p
	}
	return "(no SoundFont loaded)"
}

func (u *UI) audioText() string {
	return "48000 Hz · period 64 × 2 · polyphony 256 · per-keyboard Chorus + Delay inserts"
}
