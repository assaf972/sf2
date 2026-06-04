// Package recording captures a live performance — every MIDI event (tagged with
// its source device and a millisecond offset) plus the rendered audio — into a
// Recording that can be listed, replayed, looped and deleted.
//
// The core types are hardware-free and deterministic (a Clock is injectable), so
// the recorder, store and player are all unit-testable without an audio device
// or real wall-clock time. The audio sink and on-disk WAV writing are layered on
// top at the engine boundary; the metadata + MIDI timeline live here.
package recording

import (
	"fmt"
	"sync"
	"time"
)

// MIDIEvent is one recorded controller event, neutral of the midiio package to
// avoid an import cycle. OffsetMs is milliseconds since the take started.
type MIDIEvent struct {
	OffsetMs int64
	Kind     string // "noteon" | "noteoff" | "cc" | "pitch"
	Device   string
	A, B, C  int // note/cc/value operands (meaning depends on Kind)
}

// Recording is a finished take: metadata, the MIDI timeline, and a reference to
// the captured audio (frame count + optional WAV path written by the audio sink).
type Recording struct {
	ID            string
	Name          string
	CreatedAtUnix int64
	DurationMs    int64
	Events        []MIDIEvent
	Frames        int
	SampleRate    float64
	Samples       []float32 // captured mono audio (may be nil if capture was off)
	AudioPath     string    // set when audio is flushed to a .wav, empty otherwise
}

// HasAudio reports whether the take captured any audio frames.
func (r Recording) HasAudio() bool { return r.Frames > 0 }

// HasMIDI reports whether the take captured any MIDI events.
func (r Recording) HasMIDI() bool { return len(r.Events) > 0 }

// Clock supplies the current time in milliseconds; injectable for tests.
type Clock interface{ NowMs() int64 }

type realClock struct{}

func (realClock) NowMs() int64 { return time.Now().UnixMilli() }

// Recorder captures MIDI + audio for a single in-progress take. All methods are
// safe for concurrent use (MIDI arrives on device callbacks, audio on the render
// thread). When not armed, RecordMIDI/RecordAudio are cheap no-ops.
type Recorder struct {
	mu         sync.Mutex
	clock      Clock
	sampleRate float64

	active   bool
	name     string
	startMs  int64
	events   []MIDIEvent
	frames   int
	samples  []float32
	maxFrame int // cap on retained samples (0 = no capture)
}

// NewRecorder returns an idle recorder at the given sample rate. Pass clk=nil to
// use the real wall clock. Audio samples are retained for WAV export up to a
// 20-minute cap (frames beyond that are counted but not kept, to bound memory).
func NewRecorder(sampleRate float64, clk Clock) *Recorder {
	if clk == nil {
		clk = realClock{}
	}
	cap := 0
	if sampleRate > 0 {
		cap = int(sampleRate * 60 * 20) // ~20 minutes mono
	}
	return &Recorder{clock: clk, sampleRate: sampleRate, maxFrame: cap}
}

// IsRecording reports whether a take is currently armed.
func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active
}

// Start arms a new take, discarding any un-stopped previous capture.
func (r *Recorder) Start(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if name == "" {
		name = "Untitled take"
	}
	r.active = true
	r.name = name
	r.startMs = r.clock.NowMs()
	r.events = r.events[:0]
	r.frames = 0
	r.samples = r.samples[:0]
}

// RecordMIDI appends an event to the in-progress take (no-op when not armed).
func (r *Recorder) RecordMIDI(kind, device string, a, b, c int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		return
	}
	r.events = append(r.events, MIDIEvent{
		OffsetMs: r.clock.NowMs() - r.startMs,
		Kind:     kind, Device: device, A: a, B: b, C: c,
	})
}

// RecordAudio counts captured audio frames (no-op when not armed). The actual
// samples are streamed to the audio sink; here we track duration/length.
func (r *Recorder) RecordAudio(buf []float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		return
	}
	r.frames += len(buf)
	if r.maxFrame > 0 && len(r.samples) < r.maxFrame {
		room := r.maxFrame - len(r.samples)
		if room >= len(buf) {
			r.samples = append(r.samples, buf...)
		} else {
			r.samples = append(r.samples, buf[:room]...)
		}
	}
}

// ElapsedMs is how long the current take has been running (0 when idle).
func (r *Recorder) ElapsedMs() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		return 0
	}
	return r.clock.NowMs() - r.startMs
}

// Stop ends the take and returns the finished Recording. Returns ok=false if no
// take was armed. Duration is the longer of the MIDI timeline and the audio.
func (r *Recorder) Stop() (Recording, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		return Recording{}, false
	}
	r.active = false
	endMs := r.clock.NowMs()
	dur := endMs - r.startMs
	if r.sampleRate > 0 {
		if audioMs := int64(float64(r.frames) / r.sampleRate * 1000); audioMs > dur {
			dur = audioMs
		}
	}
	if n := len(r.events); n > 0 && r.events[n-1].OffsetMs > dur {
		dur = r.events[n-1].OffsetMs
	}
	events := make([]MIDIEvent, len(r.events))
	copy(events, r.events)
	samples := make([]float32, len(r.samples))
	copy(samples, r.samples)
	return Recording{
		Name:          r.name,
		CreatedAtUnix: r.startMs / 1000,
		DurationMs:    dur,
		Events:        events,
		Frames:        r.frames,
		SampleRate:    r.sampleRate,
		Samples:       samples,
	}, true
}

// formatID builds a stable, human-readable id from a sequence number.
func formatID(seq int) string { return fmt.Sprintf("rec-%04d", seq) }
