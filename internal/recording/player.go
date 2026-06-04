package recording

import "sync"

// PlayState is the transport state of the Player.
type PlayState int

const (
	Stopped PlayState = iota
	Playing
)

// Player replays a Recording's MIDI timeline. The host advances it by a time
// delta each tick and forwards the returned events to the synth. It supports the
// Recordings page transport: Play, Stop, and Loop. Audio playback is handled by
// the audio layer; this drives the deterministic, testable MIDI side.
type Player struct {
	mu     sync.Mutex
	rec    Recording
	loaded bool
	state  PlayState
	loop   bool
	posMs  int64
	cursor int // index of the next event to emit
}

// NewPlayer returns an idle player.
func NewPlayer() *Player { return &Player{} }

// Load sets the recording to replay and rewinds to the start (stopped).
func (p *Player) Load(r Recording) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rec = r
	p.loaded = true
	p.state = Stopped
	p.posMs = 0
	p.cursor = 0
}

// Play starts (or resumes) playback. No-op if nothing is loaded.
func (p *Player) Play() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.loaded {
		p.state = Playing
	}
}

// Stop halts playback and rewinds to the start.
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = Stopped
	p.posMs = 0
	p.cursor = 0
}

// SetLoop enables or disables looping.
func (p *Player) SetLoop(on bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loop = on
}

// Loop reports whether looping is enabled.
func (p *Player) Loop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.loop
}

// State returns the current transport state.
func (p *Player) State() PlayState {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// PositionMs returns the current playhead position.
func (p *Player) PositionMs() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.posMs
}

// Advance moves the playhead forward by dtMs and returns every event whose
// offset falls in the interval just played, in order. On reaching the end it
// either wraps (loop) — emitting the events at the wrapped start — or stops at
// the end. No-op (nil) when not playing.
func (p *Player) Advance(dtMs int64) []MIDIEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != Playing || !p.loaded || dtMs <= 0 {
		return nil
	}
	var out []MIDIEvent
	end := p.rec.DurationMs
	newPos := p.posMs + dtMs

	// Emit events from the cursor up to newPos within the current pass.
	for p.cursor < len(p.rec.Events) && p.rec.Events[p.cursor].OffsetMs <= newPos {
		out = append(out, p.rec.Events[p.cursor])
		p.cursor++
	}

	if end > 0 && newPos >= end {
		if p.loop {
			// Wrap: carry the overshoot into a fresh pass and re-emit from 0.
			over := newPos - end
			p.posMs = over
			p.cursor = 0
			for p.cursor < len(p.rec.Events) && p.rec.Events[p.cursor].OffsetMs <= over {
				out = append(out, p.rec.Events[p.cursor])
				p.cursor++
			}
		} else {
			p.posMs = end
			p.state = Stopped
		}
		return out
	}
	p.posMs = newPos
	return out
}
