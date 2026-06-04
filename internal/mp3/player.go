// Package mp3 is the backing-track player: a transport state machine
// (start/stop/loop/pitch) decoupled from the audio backend via the Output
// interface, so the logic is testable headless. The real audio backend (beep)
// implements Output; tests use a nil/fake output.
package mp3

import (
	"math"
	"sync"
)

// State is the transport state.
type State int

const (
	Stopped State = iota
	Playing
)

func (s State) String() string {
	switch s {
	case Playing:
		return "playing"
	default:
		return "stopped"
	}
}

// Output is the audio backend the Player drives. The real implementation wraps
// beep (decode + speaker + resampler + loop); tests pass nil.
type Output interface {
	Play()
	Stop()
	SetRatio(ratio float64)
	SetLoop(loop bool)
}

// Player is the transport. It owns state, loop, pitch and a logical playhead and
// forwards intent to an optional Output.
type Player struct {
	mu    sync.Mutex
	state State
	loop  bool
	pitch int     // semitones, -2..+2 in the UI
	pos   float64 // seconds
	dur   float64 // seconds
	out   Output
}

// New builds a player with an optional audio output (nil in tests).
func New(out Output) *Player { return &Player{out: out} }

// Load sets the loaded track's duration (the real backend also opens/decodes it).
func (p *Player) Load(durationSeconds float64) {
	p.mu.Lock()
	p.dur = durationSeconds
	p.pos = 0
	p.state = Stopped
	p.mu.Unlock()
}

// Start begins/resumes playback.
func (p *Player) Start() {
	p.mu.Lock()
	p.state = Playing
	out := p.out
	p.mu.Unlock()
	if out != nil {
		out.Play()
	}
}

// Stop halts playback and rewinds to the start.
func (p *Player) Stop() {
	p.mu.Lock()
	p.state = Stopped
	p.pos = 0
	out := p.out
	p.mu.Unlock()
	if out != nil {
		out.Stop()
	}
}

// ReachEnd is called when the stream hits EOF: loop restarts from 0, otherwise
// playback stops.
func (p *Player) ReachEnd() {
	p.mu.Lock()
	loop := p.loop
	p.mu.Unlock()
	if loop {
		p.mu.Lock()
		p.pos = 0
		p.state = Playing
		out := p.out
		p.mu.Unlock()
		if out != nil {
			out.Play()
		}
		return
	}
	p.Stop()
}

// SetLoop toggles looping.
func (p *Player) SetLoop(on bool) {
	p.mu.Lock()
	p.loop = on
	out := p.out
	p.mu.Unlock()
	if out != nil {
		out.SetLoop(on)
	}
}

// SetPitch sets the pitch shift in semitones and updates the resample ratio.
func (p *Player) SetPitch(semitones int) {
	p.mu.Lock()
	p.pitch = semitones
	out := p.out
	r := ratioFor(semitones)
	p.mu.Unlock()
	if out != nil {
		out.SetRatio(r)
	}
}

// State / Position / Loop / Pitch / Ratio expose transport state.
func (p *Player) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}
func (p *Player) Position() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pos
}
func (p *Player) Loop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.loop
}
func (p *Player) Pitch() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pitch
}

// Ratio is the resample ratio for the current pitch (2^(semitones/12)).
func (p *Player) Ratio() float64 { return ratioFor(p.Pitch()) }

func ratioFor(semitones int) float64 { return math.Pow(2, float64(semitones)/12.0) }
