package app

import (
	"fmt"
	"math"
	"sync"

	"gigsynth/internal/db"
	"gigsynth/internal/engine"
	"gigsynth/internal/fx"
	"gigsynth/internal/midiio"
	"gigsynth/internal/midimap"
	"gigsynth/internal/recording"
)

// Controller is the application core. It owns the synth engine and MIDI
// manager, holds the current performance Scene, and routes incoming MIDI from
// each keyboard to the right layers (with key splits and transpose).
type Controller struct {
	eng  engine.Synth
	midi *midiio.Manager

	mu      sync.RWMutex
	scene   Scene
	presets []engine.Preset
	sfPath  string

	// chains holds the per-keyboard effects insert chain (chorus -> delay).
	chains [NumLayers]*fx.Chain

	// rec captures live MIDI + audio (S21); recStore holds finished takes.
	rec      *recording.Recorder
	recStore *recording.Store

	// maps holds the per-keyboard manufacturer CC mapping.
	maps [NumLayers]midimap.Map

	// Setlist cursor (Live shuttle / Pi Zero part navigation).
	lib      *Library
	songName string
	parts    []db.Part
	partIdx  int

	// onChange is an optional UI refresh hook.
	onChange func()
}

var errNoLibrary = fmt.Errorf("controller: no library attached")

// NewController builds the controller and wires the MIDI handler. It accepts any
// engine.Synth, so tests can pass a FakeEngine with no audio device.
func NewController(eng engine.Synth, m *midiio.Manager) *Controller {
	c := &Controller{
		eng:      eng,
		midi:     m,
		scene:    DefaultScene(),
		rec:      recording.NewRecorder(engine.DefaultConfig().SampleRate, nil),
		recStore: recording.NewStore(),
	}
	for i := range c.chains {
		c.chains[i] = fx.NewChain(engine.DefaultConfig().SampleRate)
		c.maps[i] = midimap.Get("Generic GM")
	}
	// m is nil in unit tests that drive routing directly; only wire the MIDI
	// callback when a real device manager is present.
	if m != nil {
		m.SetHandler(c.handleMIDI)
	}
	return c
}

func (c *Controller) SetOnChange(f func()) { c.onChange = f }

func (c *Controller) notify() {
	if c.onChange != nil {
		c.onChange()
	}
}

// ---- SoundFont / presets ----

// LoadSoundFont loads an .sf2 and re-applies the current scene against it.
func (c *Controller) LoadSoundFont(path string) error {
	presets, err := c.eng.LoadSoundFont(path)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.presets = presets
	c.sfPath = path
	c.mu.Unlock()
	c.ApplyScene()
	c.notify()
	return nil
}

func (c *Controller) Presets() []engine.Preset {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]engine.Preset, len(c.presets))
	copy(out, c.presets)
	return out
}

func (c *Controller) SoundFontPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sfPath
}

// ---- Scene access ----

func (c *Controller) Scene() Scene {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// deep copy layers so callers can't mutate under the lock
	s := c.scene
	s.Layers = append([]Layer(nil), c.scene.Layers...)
	return s
}

// ApplyScene pushes every layer's program assignment and mix to the engine.
func (c *Controller) ApplyScene() {
	c.mu.RLock()
	layers := append([]Layer(nil), c.scene.Layers...)
	gain := c.scene.MasterGain
	hasSF := c.sfPath != ""
	c.mu.RUnlock()

	c.eng.SetMasterGain(gain)
	for i, l := range layers {
		if hasSF {
			// Ignore errors here: a scene may reference presets a given SF2
			// doesn't have; the UI surfaces preset selection explicitly.
			_ = c.eng.SelectProgram(l.Channel, l.Bank, l.Program)
		}
		vol := l.Volume
		if l.Mute || !l.Enabled {
			vol = 0
		}
		c.eng.SetChannelVolume(l.Channel, vol)
		c.eng.SetChannelPan(l.Channel, l.Pan)
		// Restore the per-keyboard effects chain for this layer.
		c.applyChorus(i, l.ChorusOn, l.ChorusRate, l.ChorusDepth)
		c.applyPhaser(i, l.PhaserOn, l.PhaserRate, l.PhaserDepth, l.PhaserFeedback)
		c.applyFlanger(i, l.FlangerOn, l.FlangerRate, l.FlangerDepth, l.FlangerFeedback, l.FlangerMix)
		c.applyDelay(i, l.DelayOn, l.DelayTime, l.DelayFeedback, l.DelayMix)
	}
}

// ---- Mutators (called from UI) ----

func (c *Controller) SetLayerPreset(idx, bank, program int, name string) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].Bank = bank
	c.scene.Layers[idx].Program = program
	c.scene.Layers[idx].PresetName = name
	ch := c.scene.Layers[idx].Channel
	c.mu.Unlock()
	_ = c.eng.SelectProgram(ch, bank, program)
	c.notify()
}

func (c *Controller) SetLayerVolume(idx, vol int) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].Volume = vol
	l := c.scene.Layers[idx]
	c.mu.Unlock()
	eff := vol
	if l.Mute || !l.Enabled {
		eff = 0
	}
	c.eng.SetChannelVolume(l.Channel, eff)
}

func (c *Controller) SetLayerPan(idx, pan int) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].Pan = pan
	ch := c.scene.Layers[idx].Channel
	c.mu.Unlock()
	c.eng.SetChannelPan(ch, pan)
}

func (c *Controller) SetLayerMute(idx int, mute bool) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].Mute = mute
	l := c.scene.Layers[idx]
	c.mu.Unlock()
	eff := l.Volume
	if mute || !l.Enabled {
		eff = 0
	}
	c.eng.SetChannelVolume(l.Channel, eff)
	c.eng.AllNotesOff(l.Channel)
	c.notify()
}

func (c *Controller) SetLayerEnabled(idx int, en bool) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].Enabled = en
	l := c.scene.Layers[idx]
	c.mu.Unlock()
	eff := l.Volume
	if l.Mute || !en {
		eff = 0
	}
	c.eng.SetChannelVolume(l.Channel, eff)
	if !en {
		c.eng.AllNotesOff(l.Channel)
	}
	c.notify()
}

func (c *Controller) SetLayerName(idx int, name string) {
	c.mu.Lock()
	if idx >= 0 && idx < len(c.scene.Layers) {
		c.scene.Layers[idx].Name = name
	}
	c.mu.Unlock()
}

func (c *Controller) SetLayerSource(idx int, source string) {
	c.mu.Lock()
	if idx >= 0 && idx < len(c.scene.Layers) {
		c.scene.Layers[idx].Source = source
	}
	c.mu.Unlock()
}

func (c *Controller) SetLayerRange(idx, low, high int) {
	c.mu.Lock()
	if idx >= 0 && idx < len(c.scene.Layers) {
		if low > high {
			low, high = high, low
		}
		c.scene.Layers[idx].KeyLow = clamp(low, 0, 127)
		c.scene.Layers[idx].KeyHigh = clamp(high, 0, 127)
	}
	c.mu.Unlock()
}

func (c *Controller) SetLayerTranspose(idx, semis int) {
	c.mu.Lock()
	if idx >= 0 && idx < len(c.scene.Layers) {
		c.scene.Layers[idx].Transpose = semis
		ch := c.scene.Layers[idx].Channel
		c.mu.Unlock()
		c.eng.AllNotesOff(ch) // avoid stuck notes when transpose changes
		return
	}
	c.mu.Unlock()
}

// SetChorus updates a keyboard's chorus state (UI/CC driven) and pushes it to
// that keyboard's DSP insert chain.
func (c *Controller) SetChorus(idx int, on bool, rate float64, depth int) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].ChorusOn = on
	c.scene.Layers[idx].ChorusRate = rate
	c.scene.Layers[idx].ChorusDepth = depth
	c.mu.Unlock()
	c.applyChorus(idx, on, rate, depth)
	c.notify()
}

// SetDelay updates a keyboard's delay state and pushes it to the DSP chain.
func (c *Controller) SetDelay(idx int, on bool, timeMs, feedback, mix int) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].DelayOn = on
	c.scene.Layers[idx].DelayTime = timeMs
	c.scene.Layers[idx].DelayFeedback = feedback
	c.scene.Layers[idx].DelayMix = mix
	c.mu.Unlock()
	c.applyDelay(idx, on, timeMs, feedback, mix)
	c.notify()
}

func (c *Controller) applyChorus(idx int, on bool, rate float64, depth int) {
	if idx < 0 || idx >= len(c.chains) || c.chains[idx] == nil {
		return
	}
	ch := c.chains[idx].Chorus
	ch.SetEnabled(on)
	ch.SetParam("rate", rate)
	ch.SetParam("depth", float64(depth))
}

// SetPhaser updates a keyboard's phaser state and pushes it to the DSP chain.
func (c *Controller) SetPhaser(idx int, on bool, rate float64, depth, feedback int) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].PhaserOn = on
	c.scene.Layers[idx].PhaserRate = rate
	c.scene.Layers[idx].PhaserDepth = depth
	c.scene.Layers[idx].PhaserFeedback = feedback
	c.mu.Unlock()
	c.applyPhaser(idx, on, rate, depth, feedback)
	c.notify()
}

// SetFlanger updates a keyboard's flanger state and pushes it to the DSP chain.
func (c *Controller) SetFlanger(idx int, on bool, rate float64, depth, feedback, mix int) {
	c.mu.Lock()
	if idx < 0 || idx >= len(c.scene.Layers) {
		c.mu.Unlock()
		return
	}
	c.scene.Layers[idx].FlangerOn = on
	c.scene.Layers[idx].FlangerRate = rate
	c.scene.Layers[idx].FlangerDepth = depth
	c.scene.Layers[idx].FlangerFeedback = feedback
	c.scene.Layers[idx].FlangerMix = mix
	c.mu.Unlock()
	c.applyFlanger(idx, on, rate, depth, feedback, mix)
	c.notify()
}

func (c *Controller) applyPhaser(idx int, on bool, rate float64, depth, feedback int) {
	if idx < 0 || idx >= len(c.chains) || c.chains[idx] == nil {
		return
	}
	p := c.chains[idx].Phaser
	p.SetEnabled(on)
	p.SetParam("rate", rate)
	p.SetParam("depth", float64(depth))
	p.SetParam("feedback", float64(feedback))
}

func (c *Controller) applyFlanger(idx int, on bool, rate float64, depth, feedback, mix int) {
	if idx < 0 || idx >= len(c.chains) || c.chains[idx] == nil {
		return
	}
	f := c.chains[idx].Flanger
	f.SetEnabled(on)
	f.SetParam("rate", rate)
	f.SetParam("depth", float64(depth))
	f.SetParam("feedback", float64(feedback))
	f.SetParam("mix", float64(mix))
}

func (c *Controller) applyDelay(idx int, on bool, timeMs, feedback, mix int) {
	if idx < 0 || idx >= len(c.chains) || c.chains[idx] == nil {
		return
	}
	d := c.chains[idx].Delay
	d.SetEnabled(on)
	d.SetParam("time", float64(timeMs))
	d.SetParam("feedback", float64(feedback))
	d.SetParam("mix", float64(mix))
}

func (c *Controller) SetMasterGain(g float64) {
	c.mu.Lock()
	c.scene.MasterGain = g
	c.mu.Unlock()
	c.eng.SetMasterGain(g)
}

// SetScene replaces the whole scene (used when recalling a saved scene) and
// applies it to the engine.
func (c *Controller) SetScene(s Scene) {
	c.mu.Lock()
	// Preserve the fixed channel mapping per slot regardless of saved data.
	for i := range s.Layers {
		s.Layers[i].Channel = i
	}
	c.scene = s
	c.mu.Unlock()
	c.Panic()
	c.ApplyScene()
	c.notify()
}

// ---- Live ----

func (c *Controller) Panic() {
	c.eng.Panic()
	// Re-apply mixer because system_reset clears CC state.
	c.ApplyScene()
}

// NoteOn/NoteOff for the on-screen test keyboard (acts as a virtual source).
func (c *Controller) VirtualNoteOn(key, vel int) {
	c.routeNote(true, "Virtual Keyboard", key, vel)
}
func (c *Controller) VirtualNoteOff(key int) {
	c.routeNote(false, "Virtual Keyboard", key, 0)
}

// ---- MIDI routing ----

func (c *Controller) handleMIDI(ev midiio.Event) {
	switch ev.Type {
	case midiio.NoteOn:
		c.rec.RecordMIDI("noteon", ev.Device, ev.Key, ev.Velocity, 0)
		c.routeNote(true, ev.Device, ev.Key, ev.Velocity)
	case midiio.NoteOff:
		c.rec.RecordMIDI("noteoff", ev.Device, ev.Key, 0, 0)
		c.routeNote(false, ev.Device, ev.Key, 0)
	case midiio.ControlChange:
		c.rec.RecordMIDI("cc", ev.Device, ev.Control, ev.Value, 0)
		c.applyCC(ev.Device, ev.Control, ev.Value)
	case midiio.PitchBend:
		c.rec.RecordMIDI("pitch", ev.Device, ev.Value, 0, 0)
		c.forwardToLayers(ev.Device, func(ch int) {
			c.eng.PitchBend(ch, ev.Value)
		})
	}
}

// ---- Recording (S21) ----

// StartRecording arms a new MIDI+audio take.
func (c *Controller) StartRecording(name string) {
	c.rec.Start(name)
	c.notify()
}

// StopRecording ends the current take and files it in the recordings store,
// returning the saved recording id (empty if nothing was recording).
func (c *Controller) StopRecording() string {
	r, ok := c.rec.Stop()
	if !ok {
		return ""
	}
	saved := c.recStore.Add(r)
	c.notify()
	return saved.ID
}

// IsRecording reports whether a take is currently being captured.
func (c *Controller) IsRecording() bool { return c.rec.IsRecording() }

// RecordingElapsedMs is how long the current take has been running.
func (c *Controller) RecordingElapsedMs() int64 { return c.rec.ElapsedMs() }

// Recordings lists saved takes (newest first) for the Recordings view.
func (c *Controller) Recordings() []recording.Recording { return c.recStore.List() }

// DeleteRecording removes one saved take by id.
func (c *Controller) DeleteRecording(id string) bool { return c.recStore.Delete(id) }

// DeleteAllRecordings clears every saved take ("Delete all").
func (c *Controller) DeleteAllRecordings() int { return c.recStore.DeleteAll() }

// CaptureAudio feeds rendered audio frames to the recorder (called from the
// audio sink when armed; a no-op otherwise).
func (c *Controller) CaptureAudio(buf []float32) { c.rec.RecordAudio(buf) }

// routeNote sends a note to every layer whose source and (for note-on) key
// range match. Note-offs ignore the range/mute so notes can never get stuck.
func (c *Controller) routeNote(on bool, device string, key, vel int) {
	c.mu.RLock()
	layers := append([]Layer(nil), c.scene.Layers...)
	c.mu.RUnlock()

	for _, l := range layers {
		if !l.Enabled {
			continue
		}
		if !sourceMatches(l.Source, device) {
			continue
		}
		tk := key + l.Transpose
		if tk < 0 || tk > 127 {
			continue
		}
		if on {
			if l.Mute || key < l.KeyLow || key > l.KeyHigh {
				continue
			}
			c.eng.NoteOn(l.Channel, tk, vel)
		} else {
			c.eng.NoteOff(l.Channel, tk)
		}
	}
}

func (c *Controller) forwardToLayers(device string, fn func(ch int)) {
	c.mu.RLock()
	layers := append([]Layer(nil), c.scene.Layers...)
	c.mu.RUnlock()
	for _, l := range layers {
		if l.Enabled && sourceMatches(l.Source, device) {
			fn(l.Channel)
		}
	}
}

// SetKeyboardMapping assigns a manufacturer mapping preset to a keyboard slot.
func (c *Controller) SetKeyboardMapping(slot int, preset string) {
	c.mu.Lock()
	if slot >= 0 && slot < len(c.maps) {
		c.maps[slot] = midimap.Get(preset)
	}
	c.mu.Unlock()
}

// applyCC resolves an incoming control change against each driven keyboard's
// mapping. Mapped CCs drive the mixer/effects; unmapped CCs pass through to the
// engine (e.g. sustain on a generic map).
func (c *Controller) applyCC(device string, cc, val int) {
	c.mu.RLock()
	type hit struct {
		idx, ch int
	}
	var hits []hit
	layers := append([]Layer(nil), c.scene.Layers...)
	maps := c.maps
	for i, l := range c.scene.Layers {
		if l.Enabled && sourceMatches(l.Source, device) {
			hits = append(hits, hit{i, l.Channel})
		}
	}
	c.mu.RUnlock()

	for _, h := range hits {
		if action, ok := maps[h.idx].Lookup(cc); ok {
			c.applyCCAction(h.idx, action, val, layers[h.idx])
		} else {
			c.eng.CC(h.ch, cc, val) // passthrough (e.g. unmapped pedals)
		}
	}
}

func ccPercent(v int) int { return int(math.Round(float64(v) * 100 / 127)) }

func (c *Controller) applyCCAction(idx int, a midimap.Action, val int, l Layer) {
	switch a {
	case midimap.Volume:
		c.SetLayerVolume(idx, val)
	case midimap.Pan:
		c.SetLayerPan(idx, val)
	case midimap.ChorusRate:
		c.SetChorus(idx, l.ChorusOn, float64(val)/127*5.0, l.ChorusDepth)
	case midimap.ChorusDepth:
		c.SetChorus(idx, l.ChorusOn, l.ChorusRate, val)
	case midimap.DelayTime:
		c.SetDelay(idx, l.DelayOn, val*1000/127, l.DelayFeedback, l.DelayMix)
	case midimap.DelayFeedback:
		c.SetDelay(idx, l.DelayOn, l.DelayTime, ccPercent(val), l.DelayMix)
	case midimap.DelayMix:
		c.SetDelay(idx, l.DelayOn, l.DelayTime, l.DelayFeedback, ccPercent(val))
	case midimap.Sustain:
		c.eng.CC(l.Channel, 64, val)
	}
}

func sourceMatches(layerSource, device string) bool {
	return layerSource == "" || layerSource == SourceAny || layerSource == device
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Helper for UI labels.
func NoteName(n int) string {
	names := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	if n < 0 || n > 127 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%s%d", names[n%12], n/12-1)
}
