package app

import (
	"encoding/json"

	"gigsynth/internal/db"
)

// settingsKey is the single row under which app settings JSON is stored.
const settingsKey = "app"

// Settings holds user-configurable application options (Settings view).
type Settings struct {
	SF2Path     string `json:"sf2Path"`
	AudioDriver string `json:"audioDriver"`
	SampleRate  int    `json:"sampleRate"`
	PeriodSize  int    `json:"periodSize"`
	Periods     int    `json:"periods"`
	Polyphony   int    `json:"polyphony"`
	MP3Folder   string `json:"mp3Folder"`
	Theme       string `json:"theme"`
	StartTab    string `json:"startTab"`
}

// DefaultSettings returns sensible defaults matching the engine config.
func DefaultSettings() Settings {
	return Settings{
		SampleRate: 48000, PeriodSize: 64, Periods: 2, Polyphony: 256,
		Theme: "Dark (stage)", StartTab: "Live",
	}
}

// LoadSettings reads settings from the database, falling back to defaults.
func LoadSettings(d *db.DB) (Settings, error) {
	s := DefaultSettings()
	raw, ok, err := d.GetSetting(settingsKey)
	if err != nil {
		return s, err
	}
	if !ok {
		return s, nil
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return DefaultSettings(), err
	}
	return s, nil
}

// SaveSettings writes settings to the database.
func SaveSettings(d *db.DB, s Settings) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return d.SetSetting(settingsKey, string(data))
}
