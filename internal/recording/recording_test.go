package recording

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClock returns scripted millisecond timestamps so recording timing is
// deterministic in tests.
type fakeClock struct {
	now  int64
	step int64
}

func (f *fakeClock) NowMs() int64 {
	t := f.now
	f.now += f.step
	return t
}

// S21-T01: the recorder timestamps MIDI relative to Start, reports IsRecording,
// and Stop returns a Recording with the captured events, frames and duration.
func TestRecorderCapturesTimelineAndAudio(t *testing.T) {
	clk := &fakeClock{now: 1_000_000, step: 100} // each clock read advances 100 ms
	r := NewRecorder(1000, clk)                  // sr=1000 => 1 frame == 1 ms

	assert.False(t, r.IsRecording())
	r.Start("Soundcheck") // reads clock -> startMs = 1_000_000 (now becomes +100)
	assert.True(t, r.IsRecording())

	r.RecordMIDI("noteon", "KB1", 60, 100, 0) // offset 100
	r.RecordMIDI("noteoff", "KB1", 60, 0, 0)  // offset 200
	r.RecordAudio(make([]float32, 500))       // 500 frames == 500 ms of audio

	rec, ok := r.Stop()
	require.True(t, ok)
	assert.False(t, r.IsRecording(), "stopped recorder is no longer armed")

	assert.Equal(t, "Soundcheck", rec.Name)
	require.Len(t, rec.Events, 2)
	assert.Equal(t, int64(100), rec.Events[0].OffsetMs)
	assert.Equal(t, "noteon", rec.Events[0].Kind)
	assert.Equal(t, int64(200), rec.Events[1].OffsetMs)
	assert.Equal(t, 500, rec.Frames)
	assert.True(t, rec.HasAudio() && rec.HasMIDI())
	// Duration is at least the audio length (500 ms).
	assert.GreaterOrEqual(t, rec.DurationMs, int64(500))

	// A second Stop with nothing armed is a safe no-op.
	_, ok = r.Stop()
	assert.False(t, ok)
}

func TestRecorderIgnoresEventsWhenIdle(t *testing.T) {
	r := NewRecorder(48000, &fakeClock{step: 1})
	r.RecordMIDI("noteon", "KB1", 60, 100, 0) // not armed
	r.RecordAudio(make([]float32, 128))
	rec, ok := r.Stop()
	assert.False(t, ok)
	assert.Empty(t, rec.Events)
}

// S21-T02: the store assigns stable IDs and supports list (newest first), get,
// delete and delete-all — the Recordings page actions.
func TestStoreCRUD(t *testing.T) {
	s := NewStore()
	a := s.Add(Recording{Name: "Take A"})
	b := s.Add(Recording{Name: "Take B"})
	assert.Equal(t, "rec-0001", a.ID)
	assert.Equal(t, "rec-0002", b.ID)

	list := s.List()
	require.Len(t, list, 2)
	assert.Equal(t, "Take B", list[0].Name, "newest first")

	got, ok := s.Get("rec-0001")
	require.True(t, ok)
	assert.Equal(t, "Take A", got.Name)

	assert.True(t, s.Delete("rec-0001"))
	assert.False(t, s.Delete("rec-0001"), "second delete is a no-op")
	assert.Equal(t, 1, s.Len())

	assert.Equal(t, 1, s.DeleteAll())
	assert.Equal(t, 0, s.Len())
}

// S21-T03: the player replays events in order as the playhead advances, stops at
// the end, and wraps cleanly when looping.
func TestPlayerReplayStopAndLoop(t *testing.T) {
	rec := Recording{
		DurationMs: 400,
		Events: []MIDIEvent{
			{OffsetMs: 100, Kind: "noteon", A: 60},
			{OffsetMs: 250, Kind: "noteoff", A: 60},
		},
	}
	p := NewPlayer()
	p.Load(rec)
	assert.Equal(t, Stopped, p.State())
	assert.Nil(t, p.Advance(1000), "advancing while stopped emits nothing")

	p.Play()
	got := p.Advance(150) // 0 -> 150 ms: first event at 100
	require.Len(t, got, 1)
	assert.Equal(t, "noteon", got[0].Kind)

	got = p.Advance(150) // 150 -> 300 ms: second event at 250
	require.Len(t, got, 1)
	assert.Equal(t, "noteoff", got[0].Kind)

	// Reach the end without loop: transport stops, playhead pinned at duration.
	got = p.Advance(200) // 300 -> 500 (>=400)
	assert.Equal(t, Stopped, p.State())
	assert.Equal(t, int64(400), p.PositionMs())

	// Now enable loop and play across the boundary: events re-emit on wrap.
	p.Stop()
	p.SetLoop(true)
	p.Play()
	p.Advance(120)       // 0 -> 120: emits the 100 ms event
	got = p.Advance(350) // 120 -> 470, wraps past 400 to 70 ms; re-emits 100? no (70<100)
	// After wrap, position is 70 ms (470-400) and we re-emitted nothing past it yet.
	assert.Equal(t, Playing, p.State(), "looping keeps playing across the end")
	assert.Equal(t, int64(70), p.PositionMs())
	// The 250 ms event was emitted before the wrap during this advance.
	assert.Contains(t, kinds(got), "noteoff")
}

func kinds(evs []MIDIEvent) []string {
	out := make([]string, len(evs))
	for i, e := range evs {
		out[i] = e.Kind
	}
	return out
}
