package app

import "gigsynth/internal/db"

// Library is the song/part service: it wraps the SQLite store and converts
// between persisted db.Part rows and the live performance Scene.
type Library struct {
	db *db.DB
}

// NewLibrary returns a Library backed by the given database.
func NewLibrary(d *db.DB) *Library { return &Library{db: d} }

// DB exposes the underlying store for direct CRUD (songs/parts lists).
func (l *Library) DB() *db.DB { return l.db }

// SceneForPart builds a recallable Scene from a stored part.
func (l *Library) SceneForPart(partID int64) (Scene, error) {
	p, err := l.db.GetPart(partID)
	if err != nil {
		return Scene{}, err
	}
	return sceneFromPart(p), nil
}

// SaveScene writes a Scene's layers (and the part name + master gain) into a part.
func (l *Library) SaveScene(partID int64, name string, s Scene) error {
	if err := l.db.SavePartMeta(partID, name, s.MasterGain); err != nil {
		return err
	}
	for i, lay := range s.Layers {
		if i >= db.NumChannels {
			break
		}
		if err := l.db.SetPartChannel(partID, channelFromLayer(i+1, lay)); err != nil {
			return err
		}
	}
	return nil
}

// ---- mapping ----

func sceneFromPart(p db.Part) Scene {
	s := Scene{Name: p.Name, MasterGain: p.MasterGain}
	s.Layers = make([]Layer, NumLayers)
	for i := 0; i < NumLayers; i++ {
		s.Layers[i] = layerFromChannel(i, p.Channels[i])
	}
	return s
}

func layerFromChannel(idx int, c db.Channel) Layer {
	src := c.Source
	if src == "" {
		src = SourceAny
	}
	return Layer{
		Channel: idx, Name: c.InstrumentName, Bank: c.Bank, Program: c.Program,
		PresetName: c.InstrumentName, Volume: c.Level, Pan: c.Pan, Mute: c.Mute, Enabled: c.Enabled,
		Source: src, KeyLow: c.KeyLow, KeyHigh: c.KeyHigh, Transpose: c.Transpose,
		ChorusOn: c.ChorusOn, ChorusRate: c.ChorusRate, ChorusDepth: c.ChorusDepth,
		DelayOn: c.DelayOn, DelayTime: c.DelayTime, DelayFeedback: c.DelayFeedback, DelayMix: c.DelayMix,
	}
}

func channelFromLayer(channel int, l Layer) db.Channel {
	return db.Channel{
		Channel: channel, Bank: l.Bank, Program: l.Program, InstrumentName: l.PresetName,
		Level: l.Volume, Pan: l.Pan, Mute: l.Mute, Enabled: l.Enabled,
		Source: l.Source, KeyLow: l.KeyLow, KeyHigh: l.KeyHigh, Transpose: l.Transpose,
		ChorusOn: l.ChorusOn, ChorusRate: l.ChorusRate, ChorusDepth: l.ChorusDepth,
		DelayOn: l.DelayOn, DelayTime: l.DelayTime, DelayFeedback: l.DelayFeedback, DelayMix: l.DelayMix,
	}
}
